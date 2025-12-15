package model

import (
	"time"

	"github.com/Shua-github/Tap-Cloud-Server/core/utils"
)

type Session struct {
	ObjectID     string `gorm:"primarykey"`
	Nickname     string
	OpenID       string
	SessionToken string `gorm:"index"`
	ShortId      string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (s Session) ToHookUser() (user utils.HookUser) {
	user.OpenID = s.OpenID
	user.SessionToken = s.SessionToken
	return
}
