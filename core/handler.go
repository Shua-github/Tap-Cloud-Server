package core

import (
	"net/http"

	"github.com/Shua-github/Tap-Cloud-Server/core/routes/custom"
	"github.com/Shua-github/Tap-Cloud-Server/core/routes/file"
	"github.com/Shua-github/Tap-Cloud-Server/core/routes/game"
	"github.com/Shua-github/Tap-Cloud-Server/core/routes/user"
	"github.com/Shua-github/Tap-Cloud-Server/core/utils"
)

type Handler struct {
	NewDb         utils.NewDb
	NewFileBucket utils.NewFileBucket
	Bucket        string
	Sign          *utils.Sign
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	db := h.NewDb(h.Bucket)
	fb := h.NewFileBucket(h.Bucket)
	if h.Sign != nil {
		custom.RegisterWhiteListRoute(mux, db, h.Sign)
	}

	file.RegisterRoutes(mux, db, h.Bucket, fb)
	user.RegisterRoutes(mux, db, h.Sign != nil)
	game.RegisterRoutes(mux, db)
}
