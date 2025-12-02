package game

import "github.com/Shua-github/Tap-Cloud-Server/core/general"

type GameSaveRequest struct {
	Summary  string          `json:"summary"`
	GameFile general.Pointer `json:"gameFile"`
	ACL      general.ACL     `json:"ACL"`
}

type GameSaveResponse struct {
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
