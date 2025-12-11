package custom

type WhiteListRequest struct {
	OpenID    string `json:"openid"`
	Timestamp uint64 `json:"timestamp"`
}
