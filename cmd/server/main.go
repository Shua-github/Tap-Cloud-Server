package main

import (
	"log"
	"net/http"

	"github.com/Shua-github/Tap-Cloud-Server/core"
	"github.com/Shua-github/Tap-Cloud-Server/core/utils"
)

func main() {
	cfg, err := LoadConfig("./config.json")
	if err != nil {
		panic(err)
	}

	var custom *utils.Custom
	if cfg.Custom.Switch {
		clinet := &http.Client{Timeout: cfg.Custom.WebHookTimeOut.ToDuration()}
		if cfg.Custom.SignKey == "" {
			panic("config miss Key")
		}
		sign := NewSign([]byte(cfg.Custom.SignKey))
		custom = &utils.Custom{Sign: sign, Client: clinet}
	}

	handler := &core.Handler{
		NewDb:         mustNewDb,
		NewFileBucket: NewLocalFileBucket,
		Bucket:        cfg.Bucket,
		Custom:        custom,
		I18nText:      &cfg.I18nText,
	}

	mux, err := handler.New()
	if err != nil {
		panic(err)
	}

	loggedMux := LoggingMiddleware(mux)

	serverAddr := "0.0.0.0:443"
	log.Printf("Server running at https://%s\n", serverAddr)
	if err := http.ListenAndServeTLS(serverAddr, cfg.Cert, cfg.Key, loggedMux); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
