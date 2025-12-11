package user

import (
	"net/http"

	"github.com/Shua-github/Tap-Cloud-Server/core/routes/custom"
	"github.com/Shua-github/Tap-Cloud-Server/core/utils"
)

func RegisterRoutes(mux *http.ServeMux, db *utils.Db, white_list bool) {
	db.AutoMigrate(&Session{}, &User{})

	mux.HandleFunc("POST /1.1/users", func(w http.ResponseWriter, r *http.Request) {
		handleRegisterUser(white_list, db, w, r)
	})
	mux.HandleFunc("PUT /1.1/users/{objectID}/refreshSessionToken", func(w http.ResponseWriter, r *http.Request) {
		handleRefreshSessionToken(db, w, r)
	})
	mux.HandleFunc("GET /1.1/users/me", func(w http.ResponseWriter, r *http.Request) {
		handleGetCurrentUser(db, w, r)
	})
	mux.HandleFunc("PUT /1.1/users/{objectID}", func(w http.ResponseWriter, r *http.Request) {
		handleUpdateUser(db, w, r)
	})
	mux.HandleFunc("DELETE /1.1/users/{objectID}", func(w http.ResponseWriter, r *http.Request) {
		handleDeleteUser(db, w, r)
	})
}

func handleRegisterUser(white_list bool, db *utils.Db, w http.ResponseWriter, r *http.Request) {
	var req TapTapRegisterUserRequest
	if err := utils.ReadJSON(r, &req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.AuthData.TapTap.OpenID == "" {
		utils.WriteError(w, http.StatusBadRequest, "missing OpenID")
		return
	}

	if white_list {
		var wl custom.WhiteList
		if err := db.Where("open_id = ?", req.AuthData.TapTap.OpenID).First(&wl).Error; err != nil {
			utils.WriteError(w, http.StatusForbidden, "OpenID not in whitelist")
			return
		}
	}

	var existing User
	if err := db.Where("open_id = ?", req.AuthData.TapTap.OpenID).First(&existing).Error; err == nil {
		var session Session
		if err := db.Where("user_object_id = ?", existing.ObjectID).First(&session).Error; err == nil {
			utils.WriteJSON(w, http.StatusOK, session.ToResp())
			return
		}
	}

	user := User{
		ObjectID: utils.RandomObjectID(),
		Nickname: req.AuthData.TapTap.Name,
		OpenID:   req.AuthData.TapTap.OpenID,
	}
	if err := db.Create(&user).Error; err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	tk := utils.RandomObjectID()
	session := Session{
		SessionToken: tk,
		UserObjectID: user.ObjectID,
	}
	if err := db.Create(&session).Error; err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "DB error")
		return
	}

	utils.WriteJSON(w, http.StatusCreated, session.ToResp())
}

func handleRefreshSessionToken(db *utils.Db, w http.ResponseWriter, r *http.Request) {
	objectID := r.PathValue("objectID")

	Session, err := GetSession(r, db)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	if Session.UserObjectID != objectID {
		utils.WriteError(w, http.StatusForbidden, "Session does not belong to this user")
		return
	}

	var user User
	if err := db.Where("object_id = ?", objectID).First(&user).Error; err != nil {
		utils.WriteError(w, http.StatusNotFound, "User not found")
		return
	}

	Session.SessionToken = utils.RandomObjectID()

	db.Save(&Session)

	utils.WriteJSON(w, http.StatusOK, Session.ToResp())
}

func handleDeleteUser(db *utils.Db, w http.ResponseWriter, r *http.Request) {
	objectID := r.PathValue("objectID")

	session, err := GetSession(r, db)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	if session.UserObjectID != objectID {
		utils.WriteError(w, http.StatusForbidden, "Cannot delete other users")
		return
	}

	db.Delete(&session)
	db.Where("object_id = ?", objectID).Delete(&User{})

	w.WriteHeader(http.StatusOK)
}

func handleGetCurrentUser(db *utils.Db, w http.ResponseWriter, r *http.Request) {
	s, err := GetSession(r, db)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	var user User
	if err := db.Where("object_id = ?", s.UserObjectID).First(&user).Error; err != nil {
		utils.WriteError(w, http.StatusNotFound, "user not found")
		return
	}

	utils.WriteJSON(w, http.StatusOK, GetCurrentUserResponse{
		ObjectID: user.ObjectID,
		Nickname: user.Nickname,
	})
}

func handleUpdateUser(db *utils.Db, w http.ResponseWriter, r *http.Request) {
	objectID := r.PathValue("objectID")
	var req UpdateUserRequest
	if err := utils.ReadJSON(r, &req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	var user User
	if err := db.Where("object_id = ?", objectID).First(&user).Error; err != nil {
		utils.WriteError(w, http.StatusNotFound, "user not found")
		return
	}

	user.Nickname = req.Nickname
	if err := db.Save(&user).Error; err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "db error")
		return
	}

	utils.WriteJSON(w, http.StatusOK, UpdateUserResponse{})
}
