package general

import (
	"encoding/json"
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
	Data string `json:"iso"`
}

func (d Date) MarshalJSON() ([]byte, error) {
	m := map[string]string{
		"__type": "Date",
		"iso":    d.Data,
	}
	return json.Marshal(m)
}

type MetaData struct {
	Size     int    `json:"size"`
	Checksum string `json:"_checksum"`
	Prefix   string `json:"prefix"`
}

type ACL map[string]map[string]bool
