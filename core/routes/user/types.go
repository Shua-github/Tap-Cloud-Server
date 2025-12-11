package user

type TapTap struct {
	OpenID string `json:"openid"`
	Name   string `json:"name"`
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

type UpdateUserResponse struct{}

type GetCurrentUserResponse struct {
	ObjectID  string `json:"objectId"`
	Nickname  string `json:"nickname"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

type SessionResponse struct {
	SessionToken string `json:"sessionToken"`
	UserObjectID string `json:"objectId"`
	CreatedAt    string `json:"createdAt"`
	UpdatedAt    string `json:"updatedAt"`
}
