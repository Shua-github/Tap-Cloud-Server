package game

import (
	"github.com/Shua-github/Tap-Cloud-Server/core/general"
)

type GameSave struct {
	Summary    string          `json:"summary"`
	GameFile   any             `json:"gameFile"`
	User       general.Pointer `json:"user"`
	CreatedAt  string          `json:"createdAt"`
	UpdatedAt  string          `json:"updatedAt"`
	ModifiedAt general.Date    `json:"modifiedAt"`
}
