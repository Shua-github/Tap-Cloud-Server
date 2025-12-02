package general

import (
	"encoding/gob"
	"encoding/json"
)

func Init() {
	gob.Register(Pointer{})
	gob.Register(File{})
}

type File struct {
	Bucket    string   `json:"bucket"`
	CreatedAt string   `json:"createdAt"`
	Key       string   `json:"key"`
	MetaData  MetaData `json:"metaData"`
	MimeType  string   `json:"mime_type"`
	Name      string   `json:"name"`
	ObjectID  string   `json:"objectId"`
	Provider  string   `json:"provider"`
	Token     string   `json:"token"`
	UploadURL string   `json:"upload_url"`
	FileURL   string   `json:"url"`
	ACL       ACL      `json:"ACL"`
}

func (f File) MarshalJSON() ([]byte, error) {
	type Alias File
	return json.Marshal(&struct {
		Type string `json:"__type"`
		Alias
	}{
		Type:  "File",
		Alias: (Alias)(f),
	})
}

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
