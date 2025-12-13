package model

import (
	"time"

	"github.com/Shua-github/Tap-Cloud-Server/core/general"
	"github.com/Shua-github/Tap-Cloud-Server/core/utils"
	"gorm.io/gorm"
)

type GameSave struct {
	Summary          string
	GameFileObjectID string
	ObjectID         string       `gorm:"primarykey"`
	ModifiedAt       general.Date `gorm:"embedded"`
	Name             string
	SessionToken     string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

func DeleteAllGameSaves(db *utils.Db, fb utils.FileBucket, sessionToken string) error {
	return db.Transaction(func(tx *gorm.DB) error {
		var saves []GameSave

		if err := tx.
			Where("session_token = ?", sessionToken).
			Select("object_id", "game_file_object_id").
			Find(&saves).Error; err != nil {
			return err
		}

		for _, save := range saves {
			ft, err := GetFile(tx, save.GameFileObjectID)
			if err != nil {
				return err
			}
			if err := ft.Delete(tx, fb); err != nil {
				return err
			}
			if err := tx.Delete(&save).Error; err != nil {
				return err
			}
		}

		return nil
	})
}
