package main

import (
	"log"
	"net/http"

	"github.com/zhazhahehedian/ai-fusion-video-go/internal/config"
	"github.com/zhazhahehedian/ai-fusion-video-go/internal/httpapi"
	"github.com/zhazhahehedian/ai-fusion-video-go/internal/platform"
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

	handler := httpapi.NewRouter(httpapi.Deps{DB: db, Redis: rdb})
	log.Printf("服务已启动，监听 %s", cfg.HTTPAddr)
	if err := http.ListenAndServe(cfg.HTTPAddr, handler); err != nil {
		log.Fatal(err)
	}
}
