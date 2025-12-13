package utils

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
)

type I18nText struct {
	OpenIDNotInWhiteList string `json:"openid_not_in_white_list"`
	ServerName           string `json:"server_name"`
}

type Sign func(data []byte) string

type Custom struct {
	Sign     Sign
	Client   *http.Client
	I18nText *I18nText
}

func (c Custom) SendWebHook(data *HookResponse, u url.URL) {
	go func() {
		body, err := json.Marshal(data)
		if err != nil {
			return
		}

		req, err := http.NewRequest(http.MethodPost, u.String(), bytes.NewBuffer(body))
		if err != nil {
			return
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Sign", c.Sign(body))

		_, _ = c.Client.Do(req)
	}()
}
