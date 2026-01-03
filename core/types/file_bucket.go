package types

import "io"

type UploadedPart struct {
	PartNumber int    `json:"partNumber"`
	ETag       string `json:"etag"`
}

type FileObject struct {
	Key  string
	Body io.ReadCloser
	ETag string
}

type MultipartUpload interface {
	UploadPart(partNumber int, data io.Reader) (UploadedPart, error)
	Complete(parts []UploadedPart) (FileObject, error)
	Abort() error
}

type FileBucket interface {
	Get(key string) (*FileObject, error)
	Delete(key string) error

	CreateMultipartUpload(key string) (uploadID string, err error)
	GetMultipartUpload(key string, uploadID string) (MultipartUpload, error)
}

type NewFileBucket func(name string) FileBucket
