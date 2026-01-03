package user

import (
	"net/http"

	"github.com/Shua-github/Tap-Cloud-Server/core/model"
	"github.com/Shua-github/Tap-Cloud-Server/core/types"
	"github.com/Shua-github/Tap-Cloud-Server/core/utils"
	"gorm.io/gorm"
)

func RegisterRoutes(mux *http.ServeMux, db *gorm.DB, c *types.Custom, fb types.FileBucket) {
	mux.HandleFunc("POST /1.1/users", func(w http.ResponseWriter, r *http.Request) {
		handleRegisterUser(c, db, w, r)
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

func handleRegisterUser(c *types.Custom, db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	var req TapTapRegisterUserRequest
	if err := utils.ReadJSON(r, &req); err != nil {
		utils.WriteError(w, types.BadRequestError)
		return
	}

	if req.AuthData.TapTap.OpenID == "" {
		utils.WriteError(w, types.BadRequestError)
		return
	}

	if c != nil {
		if err := c.UserAccessCheck(req.AuthData.TapTap.OpenID); err != nil {
			utils.WriteError(w, *err)
			return
		}
	}

	var existing model.Session
	if err := db.First(&existing, "open_id = ?", req.AuthData.TapTap.OpenID).Error; err == nil {
		if c != nil {
			c.OnEventHandler(&types.Event{
				Meta: types.EventMeta{Type: "user", Action: "login"},
				User: existing.ToEventUser(),
			})
		}
		utils.WriteJSON(w, http.StatusOK, SessionToResp(&existing))
		return
	}

	session := model.Session{
		SessionToken: utils.RandomID(),
		ObjectID:     utils.RandomID(),
		Nickname:     req.AuthData.TapTap.Name,
		OpenID:       req.AuthData.TapTap.OpenID,
		ShortId:      "Tap-Cloud-Server",
	}
	if err := db.Create(&session).Error; err != nil {
		utils.ParseDbError(w, err)
		return
	}

	if c != nil {
		c.OnEventHandler(&types.Event{
			Meta: types.EventMeta{Type: "user", Action: "create"},
			User: existing.ToEventUser(),
		})
	}

	utils.WriteJSON(w, http.StatusCreated, SessionToResp(&session))
}

func handleRefreshSessionToken(c *types.Custom, db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	objectID := r.PathValue("objectID")

	session, err := GetSession(r, db)
	if err != nil {
		utils.WriteError(w, types.BadRequestError)
		return
	}
	if session.ObjectID != objectID {
		utils.WriteError(w, types.UnauthorizedError)
		return
	}

	oldSession := *session
	session.SessionToken = utils.RandomID()

	db.Save(&session)

	if c != nil {
		c.OnEventHandler(&types.Event{
			Meta: types.EventMeta{Type: "user", Action: "refresh_session_token"},
			User: oldSession.ToEventUser(),
			Data: session.ToEventUser(),
		})
	}

	utils.WriteJSON(w, http.StatusOK, SessionToResp(session))
}

func handleDeleteUser(c *types.Custom, db *gorm.DB, fb types.FileBucket, w http.ResponseWriter, r *http.Request) {
	objectID := r.PathValue("objectID")

	session, err := GetSession(r, db)
	if err != nil {
		utils.WriteError(w, types.BadRequestError)
		return
	}
	if session.ObjectID != objectID {
		utils.WriteError(w, types.UnauthorizedError)
		return
	}

	model.DeleteAllGameSaves(db, fb, session.ObjectID)
	db.Delete(&session)

	if c != nil {
		c.OnEventHandler(&types.Event{
			Meta: types.EventMeta{Type: "user", Action: "delete"},
			User: session.ToEventUser(),
		})
	}

	w.WriteHeader(http.StatusOK)
}

func handleGetCurrentUser(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	s, err := GetSession(r, db)
	if err != nil {
		utils.ParseDbError(w, err)
		return
	}
	utils.WriteJSON(w, http.StatusOK, SessionToResp(s))
}

func handleUpdateUser(c *types.Custom, db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	objectID := r.PathValue("objectID")
	var req UpdateUserRequest
	if err := utils.ReadJSON(r, &req); err != nil {
		utils.WriteError(w, types.BadRequestError)
		return
	}

	var session model.Session
	if err := db.Where("object_id = ?", objectID).First(&session).Error; err != nil {
		utils.ParseDbError(w, err)
		return
	}

	session.Nickname = req.Nickname
	if err := db.Save(&session).Error; err != nil {
		utils.ParseDbError(w, err)
		return
	}

	if c != nil {
		c.OnEventHandler(&types.Event{
			Meta: types.EventMeta{Type: "user", Action: "update"},
			User: session.ToEventUser(),
		})

	}

	utils.WriteJSON(w, http.StatusOK, SessionToResp(&session))
}
