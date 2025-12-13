package core

import (
	"net/http"

	"github.com/Shua-github/Tap-Cloud-Server/core/model"
	"github.com/Shua-github/Tap-Cloud-Server/core/routes/custom"
	"github.com/Shua-github/Tap-Cloud-Server/core/routes/file"
	"github.com/Shua-github/Tap-Cloud-Server/core/routes/game"
	"github.com/Shua-github/Tap-Cloud-Server/core/routes/user"
	"github.com/Shua-github/Tap-Cloud-Server/core/utils"
)

type Handler struct {
	Bucket        string
	NewDb         utils.NewDb
	NewFileBucket utils.NewFileBucket
	I18nText      *utils.I18nText
	Custom        *utils.Custom
}

func (h *Handler) New() (mux *http.ServeMux, err error) {
	mux = http.NewServeMux()
	db := h.NewDb(h.Bucket)
	fb := h.NewFileBucket(h.Bucket)
	if h.Custom != nil {
		custom.RegisterWhiteListRoute(mux, db, h.Custom.Sign)
	}
	model.Init(db)
	file.RegisterRoutes(mux, db, h.Bucket, fb)
	user.RegisterRoutes(mux, db, h.Custom, h.I18nText, fb)
	game.RegisterRoutes(mux, db, h.Custom)

	return
}
