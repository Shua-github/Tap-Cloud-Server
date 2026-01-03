package model

import (
	"time"

	"github.com/Shua-github/Tap-Cloud-Server/core/types"
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

func (s Session) ToEventUser() (user types.EventUser) {
	user.OpenID = s.OpenID
	user.SessionToken = s.SessionToken
	user.Nickname = s.Nickname
	return
}
