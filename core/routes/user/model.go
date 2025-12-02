package user

type User struct {
	ObjectID  string `json:"objectId"`
	Nickname  string `json:"nickname"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
	OpenID    string `json:"-"`
}

type Session struct {
	SessionToken     string `json:"sessionToken"`
	UserObjectID     string `json:"objectId"`
	GameSaveObjectID string `json:"-"`
	CreatedAt        string `json:"createdAt"`
}
