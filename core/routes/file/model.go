package file

type UploadSession struct {
	UploadID string `json:"uploadId"`
}

type File struct {
	Data          []byte
	Key           string
	UploadID      string
	PartsReceived int
	OK            bool
}
