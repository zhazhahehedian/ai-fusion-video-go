package main

import (
	"log"
	"net/http"

	"origin.cursor.com/yuigakki-w/ai-fusion-video-study/internal/config"
	"origin.cursor.com/yuigakki-w/ai-fusion-video-study/internal/httpapi"
)

func main() {
	cfg := config.Load()
	handler := httpapi.NewRouter(cfg)

	log.Printf("phase1 healthz listening on %s (JAVA_ROOT=%s)", cfg.HTTPAddr, cfg.JavaRoot)
	if err := http.ListenAndServe(cfg.HTTPAddr, handler); err != nil {
		log.Fatal(err)
	}
}
