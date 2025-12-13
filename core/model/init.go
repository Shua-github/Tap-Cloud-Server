package model

import (
	"github.com/Shua-github/Tap-Cloud-Server/core/utils"
)

func Init(db *utils.Db) {
	db.AutoMigrate(&GameSave{})
	db.AutoMigrate(&WhiteList{})
	db.AutoMigrate(&FileToken{})
	db.AutoMigrate(&Session{})
}
