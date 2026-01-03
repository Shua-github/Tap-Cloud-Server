package types

type EventMeta struct {
	Type   string `json:"type"`
	Action string `json:"action"`
}

type EventUser struct {
	OpenID       string `json:"openid"`
	SessionToken string `json:"session_token"`
	Nickname     string `json:"nickname"`
}

type Event struct {
	Meta EventMeta `json:"meta"`
	User EventUser `json:"user"`
	Data any       `json:"data"`
}
