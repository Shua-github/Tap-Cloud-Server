package model

import "gorm.io/datatypes"

type WhiteList struct {
	OpenID  string `gorm:"primarykey"`
	WebHook datatypes.URL
}
