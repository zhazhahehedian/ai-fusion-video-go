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

	log.Printf("phase1 healthz listening on %s (JAVA_ROOT=%s)", cfg.HTTPAddr, cfg.JavaRoot)
	if err := http.ListenAndServe(cfg.HTTPAddr, handler); err != nil {
		log.Fatal(err)
	}
}
