package utils

import (
	"time"
)

type Sign struct {
	Sign func(data []byte) string
	TTL  time.Duration
}
