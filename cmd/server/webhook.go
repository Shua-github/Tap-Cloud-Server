package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"slices"

	"github.com/Shua-github/Tap-Cloud-Server/core/types"
)

func loadWhiteList(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var list []string
	if err := json.NewDecoder(file).Decode(&list); err != nil {
		return nil, err
	}
	return list, nil
}

func newWhiteListCheck(whiteList []string, msg string) types.UserAccessCheck {
	return func(openid string) *types.TCSError {
		if slices.Contains(whiteList, openid) {
			return nil
		}
		return &types.TCSError{
			HTTPCode: http.StatusForbidden,
			Message:  msg,
		}
	}
}

func newSendWebhook(key []byte, client *http.Client, webhookURL string) types.OnEventHandler {
	return func(event *types.Event) {
		go func(ev *types.Event) {
			data, err := json.Marshal(ev)
			if err != nil {
				return
			}

			signature := sign(key, data)

			req, err := http.NewRequest("POST", webhookURL, bytes.NewBuffer(data))
			if err != nil {
				return
			}
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Sign", signature)

			resp, err := client.Do(req)
			if err != nil {
				return
			}
			defer resp.Body.Close()

			_, _ = io.ReadAll(resp.Body)
		}(event)
	}
}
