package user

import (
	"errors"
	"net/http"

	"github.com/Shua-github/Tap-Cloud-Server/core/model"
	"github.com/Shua-github/Tap-Cloud-Server/core/utils"
)

func GetSession(r *http.Request, db *utils.Db) (*model.Session, error) {
	tk := utils.GetSessionToken(r)
	if tk == "" {
		return nil, errors.New("invalid request(tk)")
	}

	session := &model.Session{}
	result := db.First(session, "session_token = ?", tk)
	if result.Error != nil {
		return nil, result.Error
	}

	return session, nil
}

func SessionToResp(s *model.Session) (resp SessionResponse) {
	resp.SessionToken = s.SessionToken
	resp.ObjectID = s.ObjectID
	resp.ShortId = s.ShortId
	resp.Nickname = s.Nickname
	resp.CreatedAt = utils.FormatUTCISO(s.CreatedAt)
	resp.UpdatedAt = utils.FormatUTCISO(s.UpdatedAt)
	return
}
