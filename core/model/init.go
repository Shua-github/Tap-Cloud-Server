package model

import "gorm.io/gorm"

func Init(db *gorm.DB) {
	db.AutoMigrate(&GameSave{})
	db.AutoMigrate(&FileToken{})
	db.AutoMigrate(&Session{})
}
