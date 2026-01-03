package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/Shua-github/Tap-Cloud-Server/core/types"
)

type Config struct {
	Bucket   string         `json:"bucket"`
	Domain   string         `json:"domain"`
	Cert     string         `json:"cert"`
	Key      string         `json:"key"`
	Custom   CustomConfig   `json:"custom"`
	TapCheck TapCheckConfig `json:"tap_check"`
}

type CustomConfig struct {
	Switch               bool          `json:"switch"`
	SignKey              string        `json:"sign_key"`
	OpenIDNotInWhiteList string        `json:"openid_not_in_white_list"`
	WebHook              WebHookConfig `json:"webhook"`
}

type WebHookConfig struct {
	URL     string   `json:"url"`
	Timeout Duration `json:"timeout"`
}

type TapCheckConfig struct {
	Switch   bool   `json:"switch"`
	BaseURL  string `json:"base_url"` // open.tapapis.cn or open.tapapis.com
	ClientID string `json:"client_id"`
}

type Duration time.Duration

func (d *Duration) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	duration, err := time.ParseDuration(s)
	if err != nil {
		return fmt.Errorf("invalid duration %q: %w", s, err)
	}
	*d = Duration(duration)
	return nil
}

func LoadConfig(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	cfg := &Config{}
	if err := json.NewDecoder(file).Decode(cfg); err != nil {
		return nil, err
	}

	if cfg.Cert == "" {
		cfg.Cert = "./" + cfg.Domain + ".crt"
	}
	if cfg.Key == "" {
		cfg.Key = "./" + cfg.Domain + ".key"
	}

	return cfg, nil
}

func (d Duration) ToDuration() time.Duration {
	return time.Duration(d)
}

func initCustom(cfg *Config) (*types.Custom, error) {
	if !cfg.Custom.Switch {
		return nil, nil
	}

	if cfg.Custom.SignKey == "" {
		return nil, errors.New("custom.sign_key is required")
	}

	client := &http.Client{
		Timeout: cfg.Custom.WebHook.Timeout.ToDuration(),
	}

	whiteList, err := loadWhiteList(whiteListPath)
	if err != nil {
		return nil, err
	}

	custom := &types.Custom{
		UserAccessCheck: newWhiteListCheck(
			whiteList,
			cfg.Custom.OpenIDNotInWhiteList,
		),
		OnEventHandler: newSendWebhook(
			[]byte(cfg.Custom.SignKey),
			client,
			cfg.Custom.WebHook.URL,
		),
	}

	return custom, nil
}
