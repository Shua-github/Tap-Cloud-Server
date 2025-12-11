package user

import (
	"errors"
	"net/http"

	"github.com/Shua-github/Tap-Cloud-Server/core/utils"
	"gorm.io/gorm"
)

func GetSession(r *http.Request, db *utils.Db) (*Session, error) {
	tk := utils.GetSessionToken(r)
	if tk == "" {
		return nil, errors.New("invalid request(tk)")
	}

	session := &Session{}
	result := db.First(session, "session_token = ?", tk)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("session not found")
		}
		return nil, result.Error
	}

	return session, nil
}
