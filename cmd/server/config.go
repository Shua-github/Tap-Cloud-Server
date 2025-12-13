package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/Shua-github/Tap-Cloud-Server/core/utils"
)

type Config struct {
	Bucket   string         `json:"bucket"`
	Domain   string         `json:"domain"`
	Cert     string         `json:"cert"`
	Key      string         `json:"key"`
	I18nText utils.I18nText `json:"i18n_text"`
	Custom   CustomConfig   `json:"custom"`
}

type CustomConfig struct {
	Switch         bool     `json:"switch"`
	SignKey        string   `json:"sign_key"`
	WebHookTimeOut Duration `json:"timeout"`
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
