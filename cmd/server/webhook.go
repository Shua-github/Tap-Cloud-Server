package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"

	"github.com/Shua-github/Tap-Cloud-Server/core/utils"
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

func whiteListCheck(whiteList []string) utils.WhiteListCheck {
	return func(openid string) bool {
		for _, id := range whiteList {
			if id == openid {
				return true
			}
		}
		return false
	}
}

func sendWebhook(sign utils.Sign, client *http.Client, webhookURL string) utils.OnEventHandler {
	return func(event *utils.Event) {
		go func(ev *utils.Event) {
			data, err := json.Marshal(ev)
			if err != nil {
				return
			}

			signature := sign(data)

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
