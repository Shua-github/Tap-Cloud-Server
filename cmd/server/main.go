package main

import (
	"log"
	"net/http"

	"github.com/Shua-github/Tap-Cloud-Server/core"
)

const (
	configPath    = "./config.json"
	whiteListPath = "./white_list.json"
	serverAddr    = "0.0.0.0:443"
)

func main() {
	cfg, err := LoadConfig(configPath)
	if err != nil {
		log.Fatalf("load config failed: %v", err)
	}

	custom, err := initCustom(cfg)
	if err != nil {
		log.Fatalf("init custom failed: %v", err)
	}

	handler := &core.Handler{
		NewDb:         mustNewDb,
		NewFileBucket: NewLocalFileBucket,
		Bucket:        cfg.Bucket,
		Custom:        custom,
	}

	loggedMux := LoggingMiddleware(handler.New())

	log.Printf("Server running at https://%s", serverAddr)
	if err := http.ListenAndServeTLS(
		serverAddr,
		cfg.Cert,
		cfg.Key,
		loggedMux,
	); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
