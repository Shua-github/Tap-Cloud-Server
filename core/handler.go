package core

import (
	"net/http"

	"github.com/Shua-github/Tap-Cloud-Server/core/model"
	"github.com/Shua-github/Tap-Cloud-Server/core/routes/file"
	"github.com/Shua-github/Tap-Cloud-Server/core/routes/game"
	"github.com/Shua-github/Tap-Cloud-Server/core/routes/user"
	"github.com/Shua-github/Tap-Cloud-Server/core/types"
)

type Handler struct {
	Bucket        string
	NewDb         types.NewDb
	NewFileBucket types.NewFileBucket
	Custom        *types.Custom
}

func (h *Handler) New() (mux *http.ServeMux) {
	mux = http.NewServeMux()
	db := h.NewDb(h.Bucket)
	fb := h.NewFileBucket(h.Bucket)
	model.Init(db)
	file.RegisterRoutes(mux, db, h.Bucket, fb)
	user.RegisterRoutes(mux, db, h.Custom, fb)
	game.RegisterRoutes(mux, db, h.Custom)

	return
}
