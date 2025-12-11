package file

import (
	"encoding/json"
	"time"

	"github.com/Shua-github/Tap-Cloud-Server/core/general"
	"github.com/Shua-github/Tap-Cloud-Server/core/utils"
)

type FileToken struct {
	Bucket    string           `json:"bucket"`
	Key       string           `json:"key"`
	MetaData  general.MetaData `gorm:"embedded;embeddedPrefix:meta_" json:"metaData"`
	Name      string           `json:"name"`
	ObjectID  string           `json:"objectId" gorm:"primarykey"`
	Token     string           `json:"token"`
	UploadURL string           `json:"upload_url"`
	FileURL   string           `json:"url"`
	ACL       general.ACL      `gorm:"serializer:json" json:"ACL"`
	CreatedAt time.Time        `json:"-"`
	UpdatedAt time.Time        `json:"-"`
}

func (f FileToken) MarshalJSON() ([]byte, error) {
	type Alias FileToken
	return json.Marshal(&struct {
		Type      string `json:"__type"`
		CreatedAt string `json:"createdAt"`
		UpdatedAt string `json:"updatedAt"`
		MimeType  string `json:"mime_type"`
		Provider  string `json:"provider"`
		Alias
	}{
		Type:      "File",
		CreatedAt: utils.FormatUTCISO(f.CreatedAt),
		UpdatedAt: utils.FormatUTCISO(f.UpdatedAt),
		Provider:  "qiniu",
		MimeType:  "application/octet-stream",
		Alias:     (Alias)(f),
	})
}
