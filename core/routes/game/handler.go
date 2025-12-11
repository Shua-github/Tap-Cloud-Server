package game

import (
	"net/http"
	"time"

	"github.com/Shua-github/Tap-Cloud-Server/core/general"
	"github.com/Shua-github/Tap-Cloud-Server/core/routes/user"
	"github.com/Shua-github/Tap-Cloud-Server/core/utils"
)

func RegisterRoutes(mux *http.ServeMux, db *utils.Db) {
	db.AutoMigrate(&GameSave{})
	mux.HandleFunc("GET /1.1/classes/_GameSave", func(w http.ResponseWriter, r *http.Request) { handleGetGameSaves(db, w, r) })
	mux.HandleFunc("POST /1.1/classes/_GameSave", func(w http.ResponseWriter, r *http.Request) { handleCreateGameSave(db, w, r) })
	mux.HandleFunc("PUT /1.1/classes/_GameSave/{objectID}", func(w http.ResponseWriter, r *http.Request) { handleUpdateGameSave(db, w, r) })
}

func handleCreateGameSave(db *utils.Db, w http.ResponseWriter, r *http.Request) {
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
	session.GameSaveObjectID = utils.RandomObjectID()
	db.Save(&session)

	var game_save GameSave
	game_save.Summary = req.Summary
	game_save.GameFile = req.GameFile
	game_save.User.ObjectID = session.UserObjectID
	game_save.User.ClassName = "_User"
	game_save.User.Type = "Pointer"
	game_save.ModifiedAt = req.ModifiedAt
	game_save.Name = req.Name

	if err := db.Create(&game_save).Error; err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "DB error")
		return
	}

	utils.WriteJSON(w, http.StatusCreated, CreateGameSaveResponse{ObjectID: game_save.GameFile.ObjectID, CreatedAt: utils.FormatUTCISO(game_save.CreatedAt), UpdatedAt: utils.FormatUTCISO(game_save.UpdatedAt)})
}

func handleGetGameSaves(db *utils.Db, w http.ResponseWriter, r *http.Request) {
	var err error
	var session *user.Session
	var game_save GameSave
	var resp *GameSaveResponse
	if session, err = user.GetSession(r, db); err == nil {
		if err = db.Where("user_object_id = ?", session.UserObjectID).First(&game_save).Error; err == nil {
			if resp, err = game_save.ToResp(db); err == nil {
				utils.WriteJSON(w, http.StatusOK, resp)
				return
			}
		}
	}

	utils.WriteJSON(w, http.StatusOK, GetGameSavesResponse{Results: []any{}})
}

func handleUpdateGameSave(db *utils.Db, w http.ResponseWriter, r *http.Request) {
	objectID := r.PathValue("objectID")

	var req GameSaveRequest
	if err := utils.ReadJSON(r, &req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	var game_save GameSave
	if err := db.Where("game_file_object_id = ?", objectID).First(&game_save).Error; err != nil {
		utils.WriteError(w, http.StatusNotFound, "object not found")
		return
	}

	now := utils.FormatUTCISO(time.Now())
	game_save.Summary = req.Summary
	game_save.GameFile = req.GameFile
	game_save.ModifiedAt = general.Date{Data: now}

	if err := db.Save(&game_save).Error; err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "db error")
		return
	}

	utils.WriteJSON(w, http.StatusOK, UpdateGameSaveResponse{UpdatedAt: now})
}
