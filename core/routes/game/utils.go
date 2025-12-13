package game

import (
	"github.com/Shua-github/Tap-Cloud-Server/core/general"
	"github.com/Shua-github/Tap-Cloud-Server/core/model"
	"github.com/Shua-github/Tap-Cloud-Server/core/utils"
)

func ToCore(g model.GameSave, db *utils.Db, user_object_id string) (*GameSaveCore, error) {
	ft, err := model.GetFile(db, g.GameFileObjectID)

	if err != nil {
		return nil, err
	}
	return &GameSaveCore{
		Name:       g.Name,
		Summary:    g.Summary,
		GameFile:   *ft,
		ObjectID:   g.ObjectID,
		User:       general.Pointer{ClassName: "_User", ObjectID: user_object_id},
		ModifiedAt: g.ModifiedAt,
		CreatedAt:  utils.FormatUTCISO(g.CreatedAt),
		UpdatedAt:  utils.FormatUTCISO(g.UpdatedAt),
	}, nil
}
