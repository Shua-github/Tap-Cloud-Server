package main

import (
	"encoding/base64"

	"golang.org/x/crypto/blake2s"
)

func sign(key []byte, data []byte) string {
	h, err := blake2s.New128(key)
	if err != nil {
		panic(err.Error())
	}
	h.Write(data)
	return base64.URLEncoding.EncodeToString(h.Sum(nil))
}
