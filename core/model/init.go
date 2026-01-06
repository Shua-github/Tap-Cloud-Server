package model

import "gorm.io/gorm"

func Init(db *gorm.DB) {
	db.AutoMigrate(&GameSave{}, &FileToken{}, &Session{})
}
