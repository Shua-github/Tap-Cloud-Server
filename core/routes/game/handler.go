package game

import (
	"log"
	"net/http"

	"github.com/Shua-github/Tap-Cloud-Server/core/general"
	"github.com/Shua-github/Tap-Cloud-Server/core/routes/file"
	"github.com/Shua-github/Tap-Cloud-Server/core/routes/user"
	"github.com/Shua-github/Tap-Cloud-Server/core/utils"
)

func RegisterRoutes(mux *http.ServeMux, db utils.Db) {
	mux.HandleFunc("GET /1.1/classes/_GameSave", func(w http.ResponseWriter, r *http.Request) { handleGetGameSaves(db, w, r) })
	mux.HandleFunc("POST /1.1/classes/_GameSave", func(w http.ResponseWriter, r *http.Request) { handleCreateGameSave(db, w, r) })
	mux.HandleFunc("PUT /1.1/classes/_GameSave/{objectID}", func(w http.ResponseWriter, r *http.Request) { handleUpdateGameSave(db, w, r) })
}

func handleGetGameSaves(db utils.Db, w http.ResponseWriter, r *http.Request) {
	session, err := user.GetSession(r, db)
	if err != nil {
		utils.WriteJSON(w, http.StatusOK, GetGameSavesResponse{Results: []any{}})
		return
	}
	if err := session.Load(); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	gsm := utils.Bind(db.NewTable("gamesave_latest"), session.V.GameSaveObjectID, new(GameSave))
	if err := gsm.Load(); err == nil {
		if ptr, ok := gsm.V.GameFile.(general.Pointer); ok {
			fm := utils.Bind(db.NewTable("file"), ptr.ObjectID, new(file.File))
			if err := fm.Load(); err == nil {
				ftm := utils.Bind(db.NewTable("filetoken"), fm.V.Key, new(general.File))
				if err := ftm.Load(); err == nil {
					gsm.V.GameFile = ftm.V
				}
			}
		}
	} else {
		utils.WriteJSON(w, http.StatusOK, GetGameSavesResponse{Results: []any{}})
	}
	utils.WriteJSON(w, http.StatusOK, GetGameSavesResponse{Results: []GameSave{*gsm.V}})
}

func handleCreateGameSave(db utils.Db, w http.ResponseWriter, r *http.Request) {
	var req GameSaveRequest
	if err := utils.ReadJSON(r, &req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	session, err := user.GetSession(r, db)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := session.Load(); err != nil {
		utils.WriteError(w, http.StatusUnauthorized, err.Error())
		return
	}
	session.V.GameSaveObjectID = utils.RandomObjectID()
	if err := session.Save(); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Bb error")
		return
	}

	now := utils.GetUTCISO()
	lm := utils.Bind(db.NewTable("gamesave_latest"), session.V.GameSaveObjectID, new(GameSave))
	latest := lm.V
	latest.Summary = req.Summary
	latest.GameFile = req.GameFile
	latest.CreatedAt = now
	latest.UpdatedAt = now
	latest.User.ObjectID = session.V.UserObjectID
	latest.User.Type = "_User"
	latest.ModifiedAt = general.Date{Data: now}
	if err := lm.Save(); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusCreated, GameSaveResponse{ObjectID: session.V.GameSaveObjectID, CreatedAt: now, UpdatedAt: now})
}

func handleUpdateGameSave(db utils.Db, w http.ResponseWriter, r *http.Request) {
	objectID := r.PathValue("objectID")
	var req GameSaveRequest
	if err := utils.ReadJSON(r, &req); err != nil {
		log.Println(err.Error())
		utils.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	gslm := utils.Bind(db.NewTable("gamesave_latest"), objectID, new(GameSave))
	if err := gslm.Load(); err != nil {
		log.Println(err.Error())
		utils.WriteError(w, http.StatusNotFound, "object not found")
		return
	}
	gs := gslm.V
	gs.Summary = req.Summary
	gs.GameFile = req.GameFile
	gs.UpdatedAt = utils.GetUTCISO()
	gs.ModifiedAt = general.Date{Data: gs.UpdatedAt}
	if err := gslm.Save(); err != nil {
		log.Println(err.Error())
		utils.WriteError(w, http.StatusInternalServerError, "db error")
		return
	}

	utils.WriteJSON(w, http.StatusOK, UpdateGameSaveResponse{UpdatedAt: gs.UpdatedAt})
}
