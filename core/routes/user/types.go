package user

import "github.com/Shua-github/Tap-Cloud-Server/core/types"

type TapTap struct {
	types.ProFileInfo
	Kid    string `json:"kid"`
	MacKey string `json:"mac_key"`
}

type AuthData struct {
	TapTap TapTap `json:"taptap"`
}

type TapTapRegisterUserRequest struct {
	AuthData AuthData `json:"authData"`
}

type UpdateUserRequest struct {
	Nickname string `json:"nickname"`
}

type GetCurrentUserResponse struct {
	ObjectID  string `json:"objectId"`
	Nickname  string `json:"nickname"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

type SessionResponse struct {
	SessionToken string `json:"sessionToken"`
	ObjectID     string `json:"objectId"`
	CreatedAt    string `json:"createdAt"`
	UpdatedAt    string `json:"updatedAt"`
	Nickname     string `json:"nickname"`
	ShortId      string `json:"shortId"`
}
