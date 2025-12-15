package model

import (
	"time"

	"github.com/Shua-github/Tap-Cloud-Server/core/general"
	"github.com/Shua-github/Tap-Cloud-Server/core/utils"
)

type GameSave struct {
	Summary          string
	GameFileObjectID string       `gorm:"index"`
	ObjectID         string       `gorm:"primarykey"`
	ModifiedAt       general.Date `gorm:"embedded"`
	Name             string
	UserObjectID     string `gorm:"index"`
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

func DeleteAllGameSaves(db *utils.Db, fb utils.FileBucket, user_object_id string) error {
	var saves []GameSave

	if err := db.
		Where("user_object_id = ?", user_object_id).
		Select("object_id", "game_file_object_id").
		Find(&saves).Error; err != nil {
		return err
	}

	for _, save := range saves {
		ft, err := GetFile(db, save.GameFileObjectID)
		if err != nil {
			return err
		}
		if err := db.Delete(&save).Error; err != nil {
			return err
		}
		if err := ft.Delete(db, fb); err != nil {
			return err
		}
	}

	return nil
}
