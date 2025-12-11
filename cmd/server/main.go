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
	var sign *utils.Sign
	if cfg.Sign.Switch {
		sign = GetSign([]byte(cfg.Sign.Key), cfg.Sign.TTL.ToDuration())
	}

	handler := &core.Handler{
		NewDb:         mustNewDb,
		NewFileBucket: NewLocalFileBucket,
		Bucket:        cfg.Bucket,
		Sign:          sign,
	}

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	loggedMux := LoggingMiddleware(mux)

	serverAddr := "0.0.0.0:443"
	log.Printf("Server running at https://%s\n", serverAddr)
	if err := http.ListenAndServeTLS(serverAddr, cfg.Cert, cfg.Key, loggedMux); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
