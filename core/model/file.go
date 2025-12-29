package model

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/Shua-github/Tap-Cloud-Server/core/general"
	"github.com/Shua-github/Tap-Cloud-Server/core/utils"
	"gorm.io/gorm"
)

type FileToken struct {
	Bucket    string           `json:"bucket"`
	Key       string           `json:"key"`
	MetaData  general.MetaData `gorm:"embedded;embeddedPrefix:meta_" json:"metaData"`
	Name      string           `json:"name"`
	ObjectID  string           `json:"objectId" gorm:"primarykey"`
	Token     string           `json:"token"`
	ACL       general.ACL      `gorm:"serializer:json" json:"ACL"`
	CreatedAt time.Time        `json:"-"`
	UpdatedAt time.Time        `json:"-"`
	UploadURL string           `json:"upload_url" gorm:"-"`
	FileURL   string           `json:"url" gorm:"-"`
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

func (f *FileToken) Delete(db *gorm.DB, fb utils.FileBucket) error {
	if err := fb.Delete(f.ObjectID); err != nil {
		return fmt.Errorf("failed to delete file from bucket: %w", err)
	}

	if err := db.Delete(f).Error; err != nil {
		return fmt.Errorf("failed to delete file token record: %w", err)
	}

	return nil
}

func GetFile(db *utils.Db, ObjectID string) (*FileToken, error) {
	var ft FileToken

	if err := db.Where("object_id = ?", ObjectID).First(&ft).Error; err != nil {
		return nil, err
	}

	return &ft, nil
}
