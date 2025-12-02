package core

import (
	"net/http"

	"github.com/Shua-github/Tap-Cloud-Server/core/general"
	"github.com/Shua-github/Tap-Cloud-Server/core/routes/file"
	"github.com/Shua-github/Tap-Cloud-Server/core/routes/game"
	"github.com/Shua-github/Tap-Cloud-Server/core/routes/user"
	"github.com/Shua-github/Tap-Cloud-Server/core/utils"
)

type Handler struct {
	NewDb  utils.NewDb
	Bucket string
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	general.Init()
	db := h.NewDb(h.Bucket)

	file.RegisterRoutes(mux, db, h.Bucket)
	user.RegisterRoutes(mux, db)
	game.RegisterRoutes(mux, db)
}
