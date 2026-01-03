package types

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type ProFileInfo struct {
	OpenID string `json:"openid"`
	Name   string `json:"name"`
}

type TapError struct {
	Msg  string `json:"msg"`
	Code string `json:"error"`
}

func (e *TapError) Error() string {
	return fmt.Sprintf("TapError: code=%s, msg=%s", e.Code, e.Msg)
}

type TapResponse struct {
	Data json.RawMessage `json:"data"`
	OK   bool            `json:"success"`
}

type TapCheck struct {
	BaseURL  string // open.tapapis.cn or open.tapapis.com
	Client   *http.Client
	ClientID string
}

func (c TapCheck) GetProFileInfo(kid string, mac_key string) (*ProFileInfo, error) {
	u := &url.URL{
		Scheme: "https",
		Host:   c.BaseURL,
		Path:   "/account/profile/v1",
	}
	q := u.Query()
	q.Set("client_id", c.ClientID)
	u.RawQuery = q.Encode()

	authHeader, err := c.generateMACHeader(kid, mac_key, "GET", u)
	if err != nil {
		return nil, fmt.Errorf("failed to generate MAC: %w", err)
	}

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", authHeader)

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var tapResp TapResponse
	if err := json.Unmarshal(body, &tapResp); err != nil {
		return nil, err
	}

	if !tapResp.OK {
		var tapErr TapError
		_ = json.Unmarshal(tapResp.Data, &tapErr)
		return nil, &tapErr
	}

	var profile ProFileInfo
	if err := json.Unmarshal(tapResp.Data, &profile); err != nil {
		return nil, err
	}

	return &profile, nil
}

func (c TapCheck) generateMACHeader(kid, macKey, method string, u *url.URL) (string, error) {
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	nonce := strconv.Itoa(rand.Intn(1000000))
	port := "443"

	// RequestURI() == path + "?" + query
	pathWithQuery := u.RequestURI()

	// Format: ts + \n + nonce + \n + method + \n + path + \n + base_url + \n + port + \n\n
	signStr := fmt.Sprintf(
		"%s\n%s\n%s\n%s\n%s\n%s\n\n",
		ts,
		nonce,
		method,
		pathWithQuery,
		u.Host,
		port,
	)

	mac := hmac.New(sha1.New, []byte(macKey))
	if _, err := mac.Write([]byte(signStr)); err != nil {
		return "", err
	}

	macBase64 := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	return fmt.Sprintf(
		`MAC id="%s",ts="%s",nonce="%s",mac="%s"`,
		kid, ts, nonce, macBase64,
	), nil
}
