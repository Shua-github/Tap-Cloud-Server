package main

import (
	"encoding/base64"

	"github.com/Shua-github/Tap-Cloud-Server/core/utils"
	"golang.org/x/crypto/blake2s"
)

func NewSign(key []byte) utils.Sign {
	return func(data []byte) string {
		h, err := blake2s.New128(key)
		if err != nil {
			panic(err.Error())
		}
		h.Write(data)
		return base64.URLEncoding.EncodeToString(h.Sum(nil))
	}
}
