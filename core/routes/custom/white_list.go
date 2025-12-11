package custom

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/Shua-github/Tap-Cloud-Server/core/utils"
)

func RegisterWhiteListRoute(mux *http.ServeMux, db *utils.Db, sign *utils.Sign) {
	db.AutoMigrate(&WhiteList{})
	mux.HandleFunc("POST /add_tap_white_list", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, "Failed to read request body")
			return
		}

		sig := sign.Sign(body)
		if r.Header.Get("X-Sign") != sig {
			utils.WriteError(w, http.StatusUnauthorized, "Invalid signature")
			return
		}

		var req WhiteListRequest
		if err := json.Unmarshal(body, &req); err != nil {
			utils.WriteError(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		now := uint64(time.Now().Unix())
		if req.Timestamp+uint64(sign.TTL.Seconds()) < now {
			utils.WriteError(w, http.StatusUnauthorized, "Request expired")
			return
		}

		if err := db.Create(&WhiteList{req.OpenID}).Error; err != nil {
			utils.WriteError(w, http.StatusInternalServerError, "DB error")
			return
		}

		w.WriteHeader(http.StatusNoContent)
	})

}
