package file

import (
	"github.com/Shua-github/Tap-Cloud-Server/core/general"
	"github.com/Shua-github/Tap-Cloud-Server/core/utils"
)

type FileTokenRequest struct {
	Type     string           `json:"__type"`
	Name     string           `json:"name"`
	Prefix   string           `json:"prefix"`
	MetaData general.MetaData `json:"metaData"`
	ACL      general.ACL      `json:"ACL"`
}

type UploadPartRequest struct {
	Parts []utils.UploadedPart `json:"parts"`
}

type StartUploadResponse struct {
	UploadID string `json:"uploadId"`
}

type UploadPartResponse struct {
	Etag string `json:"etag"`
}

type CompleteUploadResponse struct {
	UploadID string `json:"uploadId"`
}

type FileCallbackResponse struct {
	Result bool `json:"result"`
}
