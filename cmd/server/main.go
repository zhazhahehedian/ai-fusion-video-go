package main

import (
	"log"
	"net/http"

	"github.com/zhazhahehedian/ai-fusion-video-go/internal/config"
	"github.com/zhazhahehedian/ai-fusion-video-go/internal/httpapi"
)

func main() {
	cfg := config.Load()
	handler := httpapi.NewRouter(cfg)

	log.Printf("服务已启动，监听 %s", cfg.HTTPAddr)
	if err := http.ListenAndServe(cfg.HTTPAddr, handler); err != nil {
		log.Fatal(err)
	}
}
