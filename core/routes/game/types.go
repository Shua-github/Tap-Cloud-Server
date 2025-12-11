package game

import (
	"github.com/Shua-github/Tap-Cloud-Server/core/general"
	"github.com/Shua-github/Tap-Cloud-Server/core/routes/file"
)

type GameSaveRequest struct {
	Summary    string          `json:"summary"`
	GameFile   general.Pointer `json:"gameFile"`
	ACL        general.ACL     `json:"ACL"`
	ModifiedAt general.Date    `json:"modifiedAt"`
	Name       string          `json:"name"`
}

type CreateGameSaveResponse struct {
	ObjectID  string `json:"objectId"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

type GetGameSavesResponse struct {
	Results any `json:"results"`
}

type UpdateGameSaveResponse struct {
	UpdatedAt string `json:"updatedAt"`
}

type GameSaveCore struct {
	Summary    string          `json:"summary"`
	GameFile   file.FileToken  `json:"gameFile"`
	User       general.Pointer `json:"user"`
	ModifiedAt general.Date    `json:"modifiedAt"`
	CreatedAt  string          `json:"createdAt"`
	UpdatedAt  string          `json:"updatedAt"`
	Name       string          `json:"name"`
}

type GameSaveResponse struct {
	Results []GameSaveCore `json:"results"`
}
