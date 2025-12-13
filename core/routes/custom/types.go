package custom

type WhiteListRequest struct {
	Exp     uint64 `json:"exp"`
	WebHook string `json:"webhook_url"`
}

type GetWhiteListResponse struct {
	OpenID  string `json:"openid"`
	WebHook string `json:"webhook_url"`
}
