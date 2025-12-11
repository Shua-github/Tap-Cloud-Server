package user

import (
	"github.com/Shua-github/Tap-Cloud-Server/core/utils"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	ObjectID string
	Nickname string
	OpenID   string
}

type Session struct {
	gorm.Model
	SessionToken     string
	UserObjectID     string
	GameSaveObjectID string
}

func (s Session) ToResp() (resp SessionResponse) {
	resp.SessionToken = s.SessionToken
	resp.UserObjectID = s.UserObjectID
	resp.CreatedAt = utils.FormatUTCISO(s.CreatedAt)
	resp.UpdatedAt = utils.FormatUTCISO(s.UpdatedAt)
	return
}
