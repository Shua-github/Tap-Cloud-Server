package model

import (
	"github.com/Shua-github/Tap-Cloud-Server/core/general"
	"github.com/Shua-github/Tap-Cloud-Server/core/types"
)

type Session struct {
	ObjectID     string `gorm:"primarykey" json:"objectId"`
	Nickname     string `json:"nickname"`
	OpenID       string `gorm:"uniqueIndex" json:"-"`
	SessionToken string `gorm:"uniqueIndex" json:"sessionToken"`
	ShortId      string `json:"shortId"`
	general.BaseDate
}

func (s Session) ToEventUser() (user types.EventUser) {
	user.OpenID = s.OpenID
	user.SessionToken = s.SessionToken
	user.Nickname = s.Nickname
	return
}
