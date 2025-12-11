package game

import (
	"github.com/Shua-github/Tap-Cloud-Server/core/general"
	"github.com/Shua-github/Tap-Cloud-Server/core/routes/file"
	"github.com/Shua-github/Tap-Cloud-Server/core/utils"
	"gorm.io/gorm"
)

type GameSave struct {
	gorm.Model
	Summary    string
	GameFile   general.Pointer `gorm:"embedded;embeddedPrefix:game_file_"`
	User       general.Pointer `gorm:"embedded;embeddedPrefix:user_"`
	ModifiedAt general.Date    `gorm:"embedded"`
	Name       string
}

func (g GameSave) ToResp(db *utils.Db) (*GameSaveResponse, error) {
	var ft file.FileToken
	if err := db.Where("object_id = ?", g.GameFile.ObjectID).First(&ft).Error; err != nil {
		return nil, err
	}

	core := GameSaveCore{
		Name:       g.Name,
		Summary:    g.Summary,
		GameFile:   ft,
		User:       g.User,
		ModifiedAt: g.ModifiedAt,
		CreatedAt:  utils.FormatUTCISO(g.CreatedAt),
		UpdatedAt:  utils.FormatUTCISO(g.UpdatedAt),
	}

	return &GameSaveResponse{
		Results: []GameSaveCore{core},
	}, nil
}
