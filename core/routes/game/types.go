package game

import (
	"github.com/Shua-github/Tap-Cloud-Server/core/general"
	"github.com/Shua-github/Tap-Cloud-Server/core/model"
)

type GameSaveRequest struct {
	Summary    string          `json:"summary"`
	GameFile   general.Pointer `json:"gameFile"`
	ACL        general.ACL     `json:"ACL"`
	ModifiedAt general.Date    `json:"modifiedAt"`
	Name       string          `json:"name"`
}

type CreateGameSaveResponse struct {
	ObjectID string `json:"objectId"`
	general.BaseDate
}

type GetGameSavesResponse struct {
	Results any `json:"results"`
}

type UpdateGameSaveResponse struct {
	general.BaseDate
}

type GameSaveCore struct {
	Summary    string          `json:"summary"`
	GameFile   model.FileToken `json:"gameFile"`
	User       general.Pointer `json:"user"`
	ModifiedAt general.Date    `json:"modifiedAt"`
	Name       string          `json:"name"`
	ObjectID   string          `json:"objectId"`
	general.BaseDate
}

type GameSaveResponse struct {
	Results []GameSaveCore `json:"results"`
}
