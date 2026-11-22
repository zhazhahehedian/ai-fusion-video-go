package main

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/zhazhahehedian/ai-fusion-video-go/internal/auth"
	"github.com/zhazhahehedian/ai-fusion-video-go/internal/config"
	"github.com/zhazhahehedian/ai-fusion-video-go/internal/generation"
	"github.com/zhazhahehedian/ai-fusion-video-go/internal/httpapi"
	"github.com/zhazhahehedian/ai-fusion-video-go/internal/logging"
	"github.com/zhazhahehedian/ai-fusion-video-go/internal/media"
	"github.com/zhazhahehedian/ai-fusion-video-go/internal/platform"
	"github.com/zhazhahehedian/ai-fusion-video-go/internal/project"
	"github.com/zhazhahehedian/ai-fusion-video-go/internal/resource"
	"github.com/zhazhahehedian/ai-fusion-video-go/internal/resource/handlers"
	"golang.org/x/sync/errgroup"
)

func main() {
	os.Exit(run(os.Stderr))
}

func run(logOut io.Writer) int {
	cfg, err := config.Load()
	if err != nil {
		slog.New(slog.NewTextHandler(logOut, &slog.HandlerOptions{Level: slog.LevelError})).
			Error("加载配置失败", "err", err)
		return 1
	}
	logger := logging.New(logOut, cfg.LogLevel, cfg.LogFormat)
	slog.SetDefault(logger)

	db, err := platform.OpenMySQL(cfg)
	if err != nil {
		logger.Error("打开 MySQL 失败", "err", err)
		return 1
	}
	defer db.Close()

	rdb, err := platform.OpenRedis(cfg)
	if err != nil {
		logger.Error("打开 Redis 失败", "err", err)
		return 1
	}
	defer rdb.Close()

	tokens := auth.NewTokenStore(rdb)
	authSvc := auth.NewService(db, tokens)
	projectSvc := project.NewService(db)

	registry, err := resource.NewRegistry(
		handlers.ProjectCreate{},
		handlers.ScriptPatch{},
		handlers.ScriptSource{},
		handlers.ScriptEpisodeCreate{},
		handlers.ScriptEpisodePatch{},
		handlers.ScriptEpisodeDelete{},
		handlers.ScriptSceneCreate{},
		handlers.ScriptScenePatch{},
		handlers.ScriptSceneDelete{},
		handlers.StoryboardCreate{},
		handlers.StoryboardPatch{},
		handlers.StoryboardClear{},
		handlers.StoryboardEpisodeClear{},
		handlers.StoryboardEpisodeCreate{},
		handlers.StoryboardEpisodePatch{},
		handlers.StoryboardEpisodeDelete{},
		handlers.StoryboardSceneCreate{},
		handlers.StoryboardScenePatch{},
		handlers.StoryboardSceneDelete{},
		handlers.StoryboardItemCreate{},
		handlers.StoryboardItemPatch{},
		handlers.StoryboardItemDelete{},
		handlers.AssetCreate{},
		handlers.AssetPatch{},
		handlers.AssetDelete{},
		handlers.AssetItemCreate{},
		handlers.AssetItemPatch{},
		handlers.AssetItemDelete{},
		handlers.ImageTaskCreate{},
		handlers.ImageTaskPatch{},
		handlers.VideoTaskCreate{},
		handlers.VideoTaskPatch{},
	)
	if err != nil {
		logger.Error("注册资源命令处理器失败", "err", err)
		return 1
	}
	commandSvc := resource.NewService(db, registry)
	scriptQuery := resource.NewScriptQuery(db)
	contentQuery := resource.NewContentQuery(db)
	generationSvc := generation.NewService(db)
	jobRepo := generation.NewRepository(db)
	outbox := generation.NewOutbox(db, logger)
	worker := generation.NewWorker(jobRepo, logger)
	mediaSvc := media.NewService(db)
	mediaRepo := media.NewRepository(db)
	mediaWorker := media.NewWorker(mediaRepo, cfg.FFmpegPath, logger)

	handler := httpapi.NewRouter(httpapi.Deps{
		Logger:     logger,
		DB:         db,
		Redis:      rdb,
		Auth:       authSvc,
		Projects:   projectSvc,
		Commands:   commandSvc,
		Scripts:    scriptQuery,
		Content:    contentQuery,
		Generation: generationSvc,
		Media:      mediaSvc,
	})

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       60 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}
	g, gctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		logger.Info("服务已启动", "addr", cfg.HTTPAddr, "env", cfg.AppEnv, "log_format", cfg.LogFormat)
		err := server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	})
	g.Go(func() error {
		<-gctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return server.Shutdown(shutdownCtx)
	})
	g.Go(func() error {
		err := outbox.Run(gctx)
		if errors.Is(err, context.Canceled) {
			return nil
		}
		return err
	})
	g.Go(func() error {
		err := worker.Run(gctx)
		if errors.Is(err, context.Canceled) {
			return nil
		}
		return err
	})
	g.Go(func() error {
		err := mediaWorker.Run(gctx)
		if errors.Is(err, context.Canceled) {
			return nil
		}
		return err
	})

	if err := g.Wait(); err != nil {
		logger.Error("进程退出", "err", err)
		return 1
	}
	return 0
}
