package utils

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func RandomObjectID() string {
	const (
		length  = 25
		charset = "0123456789abcdefghijklmnopqrstuvwxyz"
	)
	b := make([]byte, length)
	for i := range b {
		randomByte := make([]byte, 1)
		rand.Read(randomByte)
		b[i] = charset[randomByte[0]%byte(len(charset))]
	}
	return string(b)
}

func FormatUTCISO(i time.Time) string {
	return i.UTC().Format(time.RFC3339Nano)
}

func EncodeBase64Key(key string) string {
	return base64.StdEncoding.EncodeToString([]byte(key))
}

func DecodeBase64Key(encoded string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("invalid base64 key: %w", err)
	}
	return string(decoded), nil
}

func GetSessionToken(r *http.Request) string {
	return r.Header.Get("X-LC-Session")
}

func WriteJSON(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func WriteError(w http.ResponseWriter, statusCode int, message string) {
	WriteJSON(w, statusCode, map[string]any{"code": statusCode, "error": message})
}

func ReadJSON(r *http.Request, v any) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(v)
}

func CalculateMD5(data []byte) string {
	return fmt.Sprintf("%x", md5.Sum(data))
}
