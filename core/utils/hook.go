package utils

type HookMeta struct {
	Type   string `json:"type"`
	Action string `json:"action"`
}

type HookUser struct {
	OpenID       string `json:"openid"`
	SessionToken string `json:"session_token"`
}

type HookResponse struct {
	Meta HookMeta `json:"meta"`
	User HookUser `json:"user"`
	Data any      `json:"data"`
}
