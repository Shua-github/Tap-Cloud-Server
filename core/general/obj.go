package general

import (
	"encoding/json/v2"
	"time"
)

type Pointer struct {
	Type      string `json:"__type"`
	ClassName string `json:"className"`
	ObjectID  string `json:"objectId"`
}

func (p Pointer) MarshalJSON() ([]byte, error) {
	m := map[string]string{
		"__type":    "Pointer",
		"className": p.ClassName,
		"objectId":  p.ObjectID,
	}
	return json.Marshal(m)
}

type Date struct {
	Date string `json:"iso"`
}

func (d Date) MarshalJSON() ([]byte, error) {
	m := map[string]string{
		"__type": "Date",
		"iso":    d.Date,
	}
	return json.Marshal(m)
}

type MetaData struct {
	Size     int    `json:"size"`
	Checksum string `json:"_checksum"`
	Prefix   string `json:"prefix"`
}

type ACL map[string]map[string]bool

type BaseDate struct {
	CreatedAt time.Time `json:"createdAt,format:RFC3339Nano"`
	UpdatedAt time.Time `json:"updatedAt,format:RFC3339Nano"`
}
