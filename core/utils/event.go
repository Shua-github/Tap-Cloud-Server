package utils

type Meta struct {
	Type   string `json:"type"`
	Action string `json:"action"`
}

type User struct {
	OpenID       string `json:"openid"`
	SessionToken string `json:"session_token"`
	Nickname     string `json:"nickname"`
}

type Event struct {
	Meta Meta `json:"meta"`
	User User `json:"user"`
	Data any  `json:"data"`
}
