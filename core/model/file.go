package model

import (
	"encoding/json/v2"
	"fmt"

	"github.com/Shua-github/Tap-Cloud-Server/core/general"
	"github.com/Shua-github/Tap-Cloud-Server/core/types"
	"gorm.io/gorm"
)

type FileToken struct {
	Key      string           `json:"key" gorm:"uniqueIndex"`
	MetaData general.MetaData `gorm:"embedded;embeddedPrefix:meta_" json:"metaData"`
	Name     string           `json:"name"`
	ObjectID string           `json:"objectId" gorm:"primarykey"`
	Token    string           `json:"token"`
	ACL      general.ACL      `gorm:"serializer:json" json:"ACL"`

	Bucket    string `json:"bucket" gorm:"-"`
	UploadURL string `json:"upload_url,omitempty" gorm:"-"`
	FileURL   string `json:"url,omitempty" gorm:"-"`
	general.BaseDate
}

func (f FileToken) MarshalJSON() ([]byte, error) {
	type Alias FileToken
	return json.Marshal(&struct {
		Type     string `json:"__type"`
		MimeType string `json:"mime_type"`
		Provider string `json:"provider"`
		Alias
	}{
		Type:     "File",
		Provider: "qiniu",
		MimeType: "application/octet-stream",
		Alias:    (Alias)(f),
	})
}

func (f *FileToken) Delete(db *gorm.DB, fb types.FileBucket) error {
	if err := fb.Delete(f.ObjectID); err != nil {
		return fmt.Errorf("failed to delete file from bucket: %w", err)
	}

	if err := db.Delete(f).Error; err != nil {
		return fmt.Errorf("failed to delete file token record: %w", err)
	}

	return nil
}

func GetFile(db *gorm.DB, ObjectID string) (*FileToken, error) {
	var ft FileToken

	if err := db.Where("object_id = ?", ObjectID).First(&ft).Error; err != nil {
		return nil, err
	}

	return &ft, nil
}
