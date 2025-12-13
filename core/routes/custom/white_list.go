package custom

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/Shua-github/Tap-Cloud-Server/core/model"
	"github.com/Shua-github/Tap-Cloud-Server/core/utils"
	"gorm.io/datatypes"
)

func RegisterWhiteListRoute(mux *http.ServeMux, db *utils.Db, sign utils.Sign) {
	mux.HandleFunc("/custom/tap_white_list/{OpenID}", func(w http.ResponseWriter, r *http.Request) {
		OpenID := r.PathValue("OpenID")
		if OpenID == "" {
			utils.WriteError(w, http.StatusBadRequest, "OpenID is required")
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, "Failed to read request body")
			return
		}
		defer r.Body.Close()

		sig := sign(body)
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
		if req.Exp < now {
			utils.WriteError(w, http.StatusUnauthorized, "Request expired")
			return
		}

		switch r.Method {
		case http.MethodPost:
			webhook, err := url.Parse(req.WebHook)
			if err != nil {
				utils.WriteError(w, http.StatusBadRequest, "webhook is invalid")
				return
			}
			if err := db.Create(&model.WhiteList{OpenID: OpenID, WebHook: datatypes.URL(*webhook)}).Error; err != nil {
				utils.ParseDbError(w, err)
				return
			}

			w.WriteHeader(http.StatusNoContent)

		case http.MethodGet:
			var wl model.WhiteList
			if err := db.First(&wl, "open_id = ?", OpenID).Error; err != nil {
				utils.ParseDbError(w, err)
				return
			}
			utils.WriteJSON(w, http.StatusOK, GetWhiteListResponse{OpenID: wl.OpenID, WebHook: wl.WebHook.String()})

		case http.MethodPut:
			var wl model.WhiteList
			if err := db.First(&wl, "open_id = ?", OpenID).Error; err != nil {
				utils.ParseDbError(w, err)
				return
			}
			webhook, err := url.Parse(req.WebHook)
			if err != nil {
				utils.WriteError(w, http.StatusBadRequest, "webhook is invalid")
				return
			}
			wl.WebHook = datatypes.URL(*webhook)

			if err := db.Save(&wl).Error; err != nil {
				utils.WriteError(w, http.StatusInternalServerError, "DB Error:"+err.Error())
				return
			}

			w.WriteHeader(http.StatusNoContent)

		case http.MethodDelete:
			if err := db.Delete(&model.WhiteList{}, "open_id = ?", OpenID).Error; err != nil {
				utils.ParseDbError(w, err)
				return
			}
			w.WriteHeader(http.StatusNoContent)

		default:
			utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})
}
