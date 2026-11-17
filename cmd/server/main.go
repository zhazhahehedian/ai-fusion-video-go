package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/zhazhahehedian/ai-fusion-video-go/internal/auth"
	"github.com/zhazhahehedian/ai-fusion-video-go/internal/config"
	"github.com/zhazhahehedian/ai-fusion-video-go/internal/generation"
	"github.com/zhazhahehedian/ai-fusion-video-go/internal/httpapi"
	"github.com/zhazhahehedian/ai-fusion-video-go/internal/media"
	"github.com/zhazhahehedian/ai-fusion-video-go/internal/platform"
	"github.com/zhazhahehedian/ai-fusion-video-go/internal/project"
	"github.com/zhazhahehedian/ai-fusion-video-go/internal/resource"
	"github.com/zhazhahehedian/ai-fusion-video-go/internal/resource/handlers"
	"golang.org/x/sync/errgroup"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	db, err := platform.OpenMySQL(cfg)
	if err != nil {
		log.Fatalf("%v", err)
	}
	defer db.Close()

	rdb, err := platform.OpenRedis(cfg)
	if err != nil {
		log.Fatalf("%v", err)
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
	)
	if err != nil {
		log.Fatalf("注册资源命令处理器失败: %v", err)
	}
	commandSvc := resource.NewService(db, registry)
	scriptQuery := resource.NewScriptQuery(db)
	generationSvc := generation.NewService(db)
	jobRepo := generation.NewRepository(db)
	outbox := generation.NewOutbox(db)
	worker := generation.NewWorker(jobRepo)
	mediaSvc := media.NewService(db)
	mediaRepo := media.NewRepository(db)
	mediaWorker := media.NewWorker(mediaRepo, cfg.FFmpegPath)

	handler := httpapi.NewRouter(httpapi.Deps{
		DB:         db,
		Redis:      rdb,
		Auth:       authSvc,
		Projects:   projectSvc,
		Commands:   commandSvc,
		Scripts:    scriptQuery,
		Generation: generationSvc,
		Media:      mediaSvc,
	})

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	server := &http.Server{Addr: cfg.HTTPAddr, Handler: handler}
	g, gctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		log.Printf("服务已启动，监听 %s", cfg.HTTPAddr)
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
		log.Fatal(err)
	}
}
