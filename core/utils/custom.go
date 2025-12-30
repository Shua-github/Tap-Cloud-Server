package utils

import (
	"net/http"
)

type I18nText struct {
	OpenIDNotInWhiteList string `json:"openid_not_in_white_list"`
	ServerName           string `json:"server_name"`
}

type Sign func(data []byte) string
type WhiteListCheck func(openid string) bool
type OnEventHandler func(event *Event)

type Custom struct {
	Sign           Sign
	Client         *http.Client
	WhiteListCheck WhiteListCheck
	OnEventHandler OnEventHandler
	I18nText       *I18nText
}
