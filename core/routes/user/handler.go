package user

import (
	"net/http"

	"github.com/Shua-github/Tap-Cloud-Server/core/utils"
)

func RegisterRoutes(mux *http.ServeMux, db utils.Db) {
	mux.HandleFunc("POST /1.1/users", func(w http.ResponseWriter, r *http.Request) { handleRegisterUser(db, w, r) })
	mux.HandleFunc("PUT /1.1/users/{objectID}/refreshSessionToken", func(w http.ResponseWriter, r *http.Request) { handleRefreshSessionToken(db, w, r) })
	mux.HandleFunc("GET /1.1/users/me", func(w http.ResponseWriter, r *http.Request) { handleGetCurrentUser(db, w, r) })
	mux.HandleFunc("PUT /1.1/users/{objectID}", func(w http.ResponseWriter, r *http.Request) { handleUpdateUser(db, w, r) })
	mux.HandleFunc("PUT /1.1/classes/_User/{objectID}", func(w http.ResponseWriter, r *http.Request) { handleUpdateUser(db, w, r) })
	mux.HandleFunc("DELETE /1.1/users/{objectID}", func(w http.ResponseWriter, r *http.Request) { handleDeleteUser(db, w, r) })
	mux.HandleFunc("DELETE /1.1/classes/_User/{objectID}", func(w http.ResponseWriter, r *http.Request) { handleDeleteUser(db, w, r) })
}

func handleRegisterUser(db utils.Db, w http.ResponseWriter, r *http.Request) {
	var req RegisterUserRequest
	if err := utils.ReadJSON(r, &req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	openID := req.AuthData.TapTap.OpenID

	o2sTable := db.NewTable("openid2session")
	if sessionTokenBytes, err := o2sTable.Get(openID); err == nil {
		sessionToken := string(sessionTokenBytes)
		session := utils.Bind(db.NewTable("session"), sessionToken, new(Session))
		if err := session.Load(); err == nil {
			utils.WriteJSON(w, http.StatusOK, session.V)
			return
		}
		_ = o2sTable.Del(openID)
	}

	now := utils.GetUTCISO()
	user := new(User)
	user.ObjectID = utils.RandomObjectID()
	user.OpenID = openID
	user.Nickname = req.AuthData.TapTap.Name
	user.CreatedAt = now
	user.UpdatedAt = now

	userBound := utils.Bind(db.NewTable("user"), user.ObjectID, user)
	if err := userBound.Save(); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	tk := utils.RandomObjectID()
	session := utils.Bind(db.NewTable("session"), tk, new(Session))
	session.V.SessionToken = tk
	session.V.UserObjectID = user.ObjectID
	session.V.CreatedAt = now

	if err := session.Save(); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "DB error")
		return
	}

	if err := o2sTable.Put(openID, []byte(tk)); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to save openid mapping")
		return
	}

	utils.WriteJSON(w, http.StatusCreated, session.V)
}

func handleRefreshSessionToken(db utils.Db, w http.ResponseWriter, r *http.Request) {
	objectID := r.PathValue("objectID")

	oldSession, err := GetSession(r, db)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := oldSession.Load(); err != nil {
		utils.WriteError(w, http.StatusUnauthorized, "Invalid session token")
		return
	}

	if oldSession.V.UserObjectID != objectID {
		utils.WriteError(w, http.StatusForbidden, "Session does not belong to this user")
		return
	}

	user := utils.Bind(db.NewTable("user"), objectID, new(User))
	if err := user.Load(); err != nil {
		utils.WriteError(w, http.StatusNotFound, "User not found")
		return
	}

	newToken := utils.RandomObjectID()
	now := utils.GetUTCISO()

	newSession := utils.Bind(db.NewTable("session"), newToken, new(Session))
	newSession.V = oldSession.V
	newSession.V.SessionToken = newToken
	newSession.V.CreatedAt = now
	if err := newSession.Save(); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "DB error")
		return
	}
	_ = oldSession.Delete()
	o2sTable := db.NewTable("openid2session")
	if err := o2sTable.Put(user.V.OpenID, []byte(newToken)); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to update openid mapping")
		return
	}
	utils.WriteJSON(w, http.StatusOK, RefreshSessionTokenResponse{
		SessionToken: newToken,
		ObjectID:     objectID,
		CreatedAt:    newSession.V.CreatedAt,
	})
}

func handleDeleteUser(db utils.Db, w http.ResponseWriter, r *http.Request) {
	objectID := r.PathValue("objectID")

	session, err := GetSession(r, db)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := session.Load(); err != nil {
		utils.WriteError(w, http.StatusUnauthorized, "Invalid session token")
		return
	}

	if session.V.UserObjectID != objectID {
		utils.WriteError(w, http.StatusForbidden, "Cannot delete other users")
		return
	}

	user := utils.Bind(db.NewTable("user"), objectID, new(User))
	if err := user.Load(); err != nil {
		utils.WriteError(w, http.StatusNotFound, "User not found")
		return
	}

	o2sTable := db.NewTable("openid2session")
	_ = o2sTable.Del(user.V.OpenID)

	_ = session.Delete()

	_ = user.Delete()

	utils.LogResponse(http.StatusOK, "User deleted: "+objectID)

}

func handleGetCurrentUser(db utils.Db, w http.ResponseWriter, r *http.Request) {
	session, err := GetSession(r, db)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := session.Load(); err != nil {
		utils.WriteError(w, http.StatusUnauthorized, "invalid session token")
		return
	}

	um := utils.Bind(db.NewTable("user"), session.V.UserObjectID, new(User))
	if err := um.Load(); err != nil {
		utils.WriteError(w, http.StatusNotFound, "user not found")
		return
	}
	user := um.V
	utils.WriteJSON(w, http.StatusOK, GetCurrentUserResponse{ObjectID: user.ObjectID, Nickname: user.Nickname, CreatedAt: user.CreatedAt, UpdatedAt: user.UpdatedAt})
}

func handleUpdateUser(db utils.Db, w http.ResponseWriter, r *http.Request) {
	objectID := r.PathValue("objectID")
	var req UpdateUserRequest
	if err := utils.ReadJSON(r, &req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	um := utils.Bind(db.NewTable("user"), objectID, new(User))
	if err := um.Load(); err != nil {
		utils.WriteError(w, http.StatusNotFound, "user not found")
		return
	}

	um.V.Nickname = req.Nickname
	um.V.UpdatedAt = utils.GetUTCISO()
	if err := um.Save(); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "db error")
		return
	}

	utils.WriteJSON(w, http.StatusOK, UpdateUserResponse{})
}
