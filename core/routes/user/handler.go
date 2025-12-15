package user

import (
	"net/http"
	"net/url"

	"github.com/Shua-github/Tap-Cloud-Server/core/model"
	"github.com/Shua-github/Tap-Cloud-Server/core/utils"
)

func RegisterRoutes(mux *http.ServeMux, db *utils.Db, c *utils.Custom, t *utils.I18nText, fb utils.FileBucket) {
	mux.HandleFunc("POST /1.1/users", func(w http.ResponseWriter, r *http.Request) {
		handleRegisterUser(c, t, db, w, r)
	})
	mux.HandleFunc("PUT /1.1/users/{objectID}/refreshSessionToken", func(w http.ResponseWriter, r *http.Request) {
		handleRefreshSessionToken(c, db, w, r)
	})
	mux.HandleFunc("GET /1.1/users/me", func(w http.ResponseWriter, r *http.Request) {
		handleGetCurrentUser(db, w, r)
	})
	mux.HandleFunc("PUT /1.1/users/{objectID}", func(w http.ResponseWriter, r *http.Request) {
		handleUpdateUser(c, db, w, r)
	})
	mux.HandleFunc("PUT /1.1/classes/_User/{objectID}", func(w http.ResponseWriter, r *http.Request) {
		handleUpdateUser(c, db, w, r)
	})
	mux.HandleFunc("DELETE /1.1/users/{objectID}", func(w http.ResponseWriter, r *http.Request) {
		handleDeleteUser(c, db, fb, w, r)
	})
}

func handleRegisterUser(c *utils.Custom, t *utils.I18nText, db *utils.Db, w http.ResponseWriter, r *http.Request) {
	var req TapTapRegisterUserRequest
	if err := utils.ReadJSON(r, &req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.AuthData.TapTap.OpenID == "" {
		utils.WriteError(w, http.StatusBadRequest, "missing OpenID")
		return
	}

	if c != nil {
		var wl model.WhiteList
		if err := db.First(&wl, "open_id = ?", req.AuthData.TapTap.OpenID).Error; err != nil {
			utils.WriteError(w, http.StatusForbidden, t.OpenIDNotInWhiteList)
			return
		}
	}

	var existing model.Session
	if err := db.First(&existing, "open_id = ?", req.AuthData.TapTap.OpenID).Error; err == nil {
		utils.WriteJSON(w, http.StatusOK, SessionToResp(&existing))
		return
	}

	session := model.Session{
		SessionToken: utils.RandomObjectID(),
		ObjectID:     utils.RandomObjectID(),
		Nickname:     req.AuthData.TapTap.Name,
		OpenID:       req.AuthData.TapTap.OpenID,
		ShortId:      t.ServerName,
	}
	if err := db.Create(&session).Error; err != nil {
		utils.ParseDbError(w, err)
		return
	}

	utils.WriteJSON(w, http.StatusCreated, SessionToResp(&session))
}

func handleRefreshSessionToken(c *utils.Custom, db *utils.Db, w http.ResponseWriter, r *http.Request) {
	objectID := r.PathValue("objectID")

	session, err := GetSession(r, db)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	if session.ObjectID != objectID {
		utils.WriteError(w, http.StatusForbidden, "Session does not belong to this user")
		return
	}

	oldSession := *session
	session.SessionToken = utils.RandomObjectID()

	db.Save(&session)

	if c != nil {
		var wl model.WhiteList
		if err := db.First(&wl, "open_id = ?", session.OpenID).Error; err != nil {
			utils.ParseDbError(w, err)
			return
		}
		if wl.WebHook != nil {
			c.SendWebHook(&utils.HookResponse{
				Meta: utils.HookMeta{Type: "user", Action: "refresh_session_token"},
				User: oldSession.ToHookUser(),
				Data: session.ToHookUser(),
			}, url.URL(*wl.WebHook))
		}
	}

	utils.WriteJSON(w, http.StatusOK, SessionToResp(session))
}

func handleDeleteUser(c *utils.Custom, db *utils.Db, fb utils.FileBucket, w http.ResponseWriter, r *http.Request) {
	objectID := r.PathValue("objectID")

	session, err := GetSession(r, db)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	if session.ObjectID != objectID {
		utils.WriteError(w, http.StatusForbidden, "Cannot delete other users")
		return
	}

	model.DeleteAllGameSaves(db, fb, session.ObjectID)
	db.Delete(&session)

	if c != nil {
		var wl model.WhiteList
		if err := db.First(&wl, "open_id = ?", session.OpenID).Error; err != nil {
			utils.ParseDbError(w, err)
			return
		}
		if wl.WebHook != nil {
			c.SendWebHook(&utils.HookResponse{
				Meta: utils.HookMeta{Type: "user", Action: "delete"},
				User: session.ToHookUser(),
				Data: nil,
			}, url.URL(*wl.WebHook))
		}
	}

	w.WriteHeader(http.StatusOK)
}

func handleGetCurrentUser(db *utils.Db, w http.ResponseWriter, r *http.Request) {
	s, err := GetSession(r, db)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	utils.WriteJSON(w, http.StatusOK, SessionToResp(s))
}

func handleUpdateUser(c *utils.Custom, db *utils.Db, w http.ResponseWriter, r *http.Request) {
	objectID := r.PathValue("objectID")
	var req UpdateUserRequest
	if err := utils.ReadJSON(r, &req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	var session model.Session
	if err := db.Where("object_id = ?", objectID).First(&session).Error; err != nil {
		utils.WriteError(w, http.StatusNotFound, "user not found")
		return
	}

	session.Nickname = req.Nickname
	if err := db.Save(&session).Error; err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "DB Error:"+err.Error())
		return
	}

	if c != nil {
		var wl model.WhiteList
		if err := db.First(&wl, "open_id = ?", session.OpenID).Error; err != nil {
			utils.ParseDbError(w, err)
			return
		}
		if wl.WebHook != nil {
			c.SendWebHook(&utils.HookResponse{
				Meta: utils.HookMeta{Type: "user", Action: "update"},
				User: session.ToHookUser(),
				Data: HookData{session.Nickname},
			}, url.URL(*wl.WebHook))
		}
	}

	utils.WriteJSON(w, http.StatusOK, SessionToResp(&session))
}
