package game

import (
	"net/http"
	"net/url"

	"github.com/Shua-github/Tap-Cloud-Server/core/general"
	"github.com/Shua-github/Tap-Cloud-Server/core/model"
	"github.com/Shua-github/Tap-Cloud-Server/core/routes/user"
	"github.com/Shua-github/Tap-Cloud-Server/core/utils"
)

func RegisterRoutes(mux *http.ServeMux, db *utils.Db, custom *utils.Custom) {
	mux.HandleFunc("GET /1.1/classes/_GameSave", func(w http.ResponseWriter, r *http.Request) { handleGetGameSaves(db, w, r) })
	mux.HandleFunc("POST /1.1/classes/_GameSave", func(w http.ResponseWriter, r *http.Request) { handleCreateGameSave(custom, db, w, r) })
	mux.HandleFunc("PUT /1.1/classes/_GameSave/{objectID}", func(w http.ResponseWriter, r *http.Request) { handleUpdateGameSave(custom, db, w, r) })
}

func handleCreateGameSave(c *utils.Custom, db *utils.Db, w http.ResponseWriter, r *http.Request) {
	var req GameSaveRequest
	if err := utils.ReadJSON(r, &req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	session, err := user.GetSession(r, db)
	if err != nil {
		utils.ParseDbError(w, err)
		return
	}

	game_save := model.GameSave{
		Summary:          req.Summary,
		GameFileObjectID: req.GameFile.ObjectID,
		SessionToken:     session.SessionToken,
		ModifiedAt:       req.ModifiedAt,
		Name:             req.Name,
		ObjectID:         utils.RandomObjectID(),
	}

	if err := db.Create(&game_save).Error; err != nil {
		utils.ParseDbError(w, err)
		return
	}

	if c != nil {
		var wl model.WhiteList
		if err := db.First(&wl, "open_id = ?", session.OpenID).Error; err != nil {
			utils.ParseDbError(w, err)
			return
		}
		if err := db.First(&game_save, "object_id = ?", game_save.ObjectID).Error; err != nil {
			utils.ParseDbError(w, err)
			return
		}
		file, err := model.GetFile(db, game_save.GameFileObjectID)
		if err != nil {
			utils.ParseDbError(w, err)
			return
		}
		c.SendWebHook(&utils.HookResponse{
			Meta: utils.HookMeta{Type: "save", Action: "create"},
			User: session.ToHookUser(),
			Data: HookData{file.FileURL.Path, req.Summary},
		}, url.URL(wl.WebHook))
	}

	utils.WriteJSON(w, http.StatusCreated, CreateGameSaveResponse{
		ObjectID:  game_save.ObjectID,
		CreatedAt: utils.FormatUTCISO(game_save.CreatedAt),
		UpdatedAt: utils.FormatUTCISO(game_save.UpdatedAt),
	})
}

func handleGetGameSaves(db *utils.Db, w http.ResponseWriter, r *http.Request) {
	var err error
	var session *model.Session
	var game_save model.GameSave
	var resp GameSaveResponse
	if session, err = user.GetSession(r, db); err == nil {
		if err = db.First(&game_save, "session_token = ?", session.SessionToken).Error; err == nil {
			ft, err := model.GetFile(db, game_save.GameFileObjectID)
			if err != nil {
				utils.ParseDbError(w, err)
			}
			data := &GameSaveCore{
				Name:       game_save.Name,
				Summary:    game_save.Summary,
				GameFile:   *ft,
				ObjectID:   game_save.ObjectID,
				User:       general.Pointer{ClassName: "_User", ObjectID: session.ObjectID},
				ModifiedAt: game_save.ModifiedAt,
				CreatedAt:  utils.FormatUTCISO(game_save.CreatedAt),
				UpdatedAt:  utils.FormatUTCISO(game_save.UpdatedAt),
			}
			resp.Results = append(resp.Results, *data)
			utils.WriteJSON(w, http.StatusOK, resp)
			return
		}
	}

	utils.WriteJSON(w, http.StatusOK, resp)
}

func handleUpdateGameSave(c *utils.Custom, db *utils.Db, w http.ResponseWriter, r *http.Request) {
	objectID := r.PathValue("objectID")

	var req GameSaveRequest
	if err := utils.ReadJSON(r, &req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	var game_save model.GameSave
	if err := db.First(&game_save, "object_id = ?", objectID).Error; err != nil {
		utils.ParseDbError(w, err)
		return
	}

	game_save.Summary = req.Summary
	game_save.GameFileObjectID = req.GameFile.ObjectID
	game_save.ModifiedAt = req.ModifiedAt

	if err := db.Save(&game_save).Error; err != nil {
		utils.ParseDbError(w, err)
		return
	}

	if c != nil {
		var session model.Session
		if err := db.First(&session, "session_token", game_save.SessionToken).Error; err != nil {
			utils.ParseDbError(w, err)
			return
		}

		var wl model.WhiteList
		if err := db.First(&wl, "open_id = ?", session.OpenID).Error; err != nil {
			utils.ParseDbError(w, err)
			return
		}

		file, err := model.GetFile(db, game_save.GameFileObjectID)
		if err != nil {
			utils.ParseDbError(w, err)
			return
		}
		c.SendWebHook(&utils.HookResponse{
			Meta: utils.HookMeta{Type: "save", Action: "update"},
			User: session.ToHookUser(),
			Data: HookData{file.FileURL.Path, req.Summary},
		}, url.URL(wl.WebHook))
	}
	utils.WriteJSON(w, http.StatusOK, UpdateGameSaveResponse{UpdatedAt: utils.FormatUTCISO(game_save.UpdatedAt)})
}
