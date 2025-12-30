package game

import (
	"net/http"
	"net/url"

	"github.com/Shua-github/Tap-Cloud-Server/core/general"
	"github.com/Shua-github/Tap-Cloud-Server/core/model"
	"github.com/Shua-github/Tap-Cloud-Server/core/routes/user"
	"github.com/Shua-github/Tap-Cloud-Server/core/types"
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
		utils.WriteError(w, types.BadRequestError)
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
		UserObjectID:     session.ObjectID,
		ModifiedAt:       req.ModifiedAt,
		Name:             req.Name,
		ObjectID:         utils.RandomObjectID(),
	}

	if err := db.Create(&game_save).Error; err != nil {
		utils.ParseDbError(w, err)
		return
	}

	if c != nil {
		c.OnEventHandler(&utils.Event{
			Meta: utils.Meta{Type: "save", Action: "create"},
			User: session.ToUser(),
			Data: HookData{game_save.GameFileObjectID, req.Summary},
		})
	}

	utils.WriteJSON(w, http.StatusCreated, CreateGameSaveResponse{
		ObjectID:  game_save.ObjectID,
		CreatedAt: utils.FormatUTCISO(game_save.CreatedAt),
		UpdatedAt: utils.FormatUTCISO(game_save.UpdatedAt),
	})
}

func handleGetGameSaves(db *utils.Db, w http.ResponseWriter, r *http.Request) {

	var scheme string
	if r.TLS != nil {
		scheme = "https"
	} else {
		scheme = "http"
	}

	var resp GameSaveResponse

	session, err := user.GetSession(r, db)
	if err != nil {
		utils.WriteJSON(w, http.StatusOK, resp)
		return
	}

	var game_saves []model.GameSave
	if err := db.Where("user_object_id = ?", session.ObjectID).Find(&game_saves).Error; err != nil {
		utils.ParseDbError(w, err)
		return
	}

	for _, gs := range game_saves {
		ft, err := model.GetFile(db, gs.GameFileObjectID)
		if err != nil {
			utils.ParseDbError(w, err)
			return
		}
		FileURL := url.URL{
			Scheme: scheme,
			Host:   r.Host,
			Path:   "/1.1/files/" + ft.ObjectID,
		}

		ft.FileURL = FileURL.String()
		data := &GameSaveCore{
			Name:       gs.Name,
			Summary:    gs.Summary,
			GameFile:   *ft,
			ObjectID:   gs.ObjectID,
			User:       general.Pointer{ClassName: "_User", ObjectID: session.ObjectID},
			ModifiedAt: gs.ModifiedAt,
			CreatedAt:  utils.FormatUTCISO(gs.CreatedAt),
			UpdatedAt:  utils.FormatUTCISO(gs.UpdatedAt),
		}
		resp.Results = append(resp.Results, *data)
	}

	utils.WriteJSON(w, http.StatusOK, resp)
}

func handleUpdateGameSave(c *utils.Custom, db *utils.Db, w http.ResponseWriter, r *http.Request) {
	objectID := r.PathValue("objectID")

	var req GameSaveRequest
	if err := utils.ReadJSON(r, &req); err != nil {
		utils.WriteError(w, types.BadRequestError)
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
		if err := db.First(&session, "object_id = ?", game_save.UserObjectID).Error; err != nil {
			utils.ParseDbError(w, err)
			return
		}
		c.OnEventHandler(&utils.Event{
			Meta: utils.Meta{Type: "save", Action: "update"},
			User: session.ToUser(),
			Data: HookData{game_save.GameFileObjectID, req.Summary},
		})
	}

	utils.WriteJSON(w, http.StatusOK, UpdateGameSaveResponse{UpdatedAt: utils.FormatUTCISO(game_save.UpdatedAt)})
}
