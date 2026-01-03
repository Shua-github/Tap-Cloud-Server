package file

import (
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/Shua-github/Tap-Cloud-Server/core/model"
	"github.com/Shua-github/Tap-Cloud-Server/core/types"
	"github.com/Shua-github/Tap-Cloud-Server/core/utils"
	"gorm.io/gorm"
)

func RegisterRoutes(mux *http.ServeMux, db *gorm.DB, bucket string, fb types.FileBucket) {
	mux.HandleFunc("POST /1.1/fileTokens", func(w http.ResponseWriter, r *http.Request) { handleCreateFileToken(db, bucket, w, r) })
	mux.HandleFunc("DELETE /1.1/files/{ObjectID}", func(w http.ResponseWriter, r *http.Request) { handleDeleteFile(db, fb, w, r) })
	mux.HandleFunc("GET /1.1/files/{ObjectID}", func(w http.ResponseWriter, r *http.Request) { handleGetFile(fb, w, r) })

	mux.HandleFunc("POST /1.1/fileCallback", handleFileCallback)

	mux.HandleFunc("POST /buckets/{bucket}/objects/{tokenKey}/uploads", func(w http.ResponseWriter, r *http.Request) { handleStartUpload(db, fb, w, r) })
	mux.HandleFunc("PUT /buckets/{bucket}/objects/{tokenKey}/uploads/{uploadID}/{partNum}", func(w http.ResponseWriter, r *http.Request) { handleUploadPart(db, fb, w, r) })
	mux.HandleFunc("POST /buckets/{bucket}/objects/{tokenKey}/uploads/{uploadID}", func(w http.ResponseWriter, r *http.Request) { handleCompleteUpload(db, fb, w, r) })
}

func handleCreateFileToken(db *gorm.DB, bucket string, w http.ResponseWriter, r *http.Request) {
	var req FileTokenRequest
	if err := utils.ReadJSON(r, &req); err != nil {
		utils.WriteError(w, types.BadRequestError)
		return
	}

	sharedID := utils.RandomID()

	fileToken := utils.RandomID()

	var scheme string
	if r.TLS != nil {
		scheme = "https"
	} else {
		scheme = "http"
	}

	baseURL := url.URL{
		Scheme: scheme,
		Host:   r.Host,
	}

	ft := new(model.FileToken)

	ft.Bucket = bucket
	ft.Key = sharedID      // Key = ObjectID
	ft.ObjectID = sharedID // ObjectID = Key
	ft.MetaData = req.MetaData
	ft.Name = req.Name
	ft.Token = fileToken
	ft.ACL = req.ACL
	ft.UploadURL = baseURL.String()

	if err := db.Create(&ft).Error; err != nil {
		utils.ParseDbError(w, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	utils.WriteJSON(w, http.StatusOK, ft)
}

func handleGetFile(fb types.FileBucket, w http.ResponseWriter, r *http.Request) {
	objectID := r.PathValue("ObjectID")

	fileObj, err := fb.Get(objectID)
	if err != nil {
		utils.WriteError(w, types.NotFoundError)
		return
	}
	defer fileObj.Body.Close()

	w.Header().Set("Content-Type", "application/octet-stream")
	w.WriteHeader(http.StatusOK)

	_, err = io.Copy(w, fileObj.Body)
	if err != nil {
		utils.WriteError(w, types.NewUnknownError("Failed to send file"))
		return
	}
}

func handleDeleteFile(db *gorm.DB, fb types.FileBucket, w http.ResponseWriter, r *http.Request) {
	objectID := r.PathValue("ObjectID")
	var file model.FileToken

	if err := db.First(&file, "object_id = ?", objectID).Error; err != nil {
		utils.ParseDbError(w, err)
		return
	}
	if err := file.Delete(db, fb); err != nil {
		utils.ParseDbError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func handleFileCallback(w http.ResponseWriter, r *http.Request) {
	utils.WriteJSON(w, http.StatusOK, FileCallbackResponse{Result: true})
}

func handleStartUpload(db *gorm.DB, fb types.FileBucket, w http.ResponseWriter, r *http.Request) {
	key, _ := utils.DecodeBase64Key(r.PathValue("tokenKey"))

	var ft model.FileToken
	if err := db.Where("key = ?", key).First(&ft).Error; err != nil {
		utils.ParseDbError(w, err)
		return
	}

	uploadID, err := fb.CreateMultipartUpload(ft.Key)
	if err != nil {
		utils.ParseDbError(w, err)
		return

	}

	ft.Token = utils.RandomID()
	if err := db.Save(&ft).Error; err != nil {
		utils.ParseDbError(w, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, StartUploadResponse{UploadID: uploadID})
}

func handleUploadPart(db *gorm.DB, fb types.FileBucket, w http.ResponseWriter, r *http.Request) {
	uploadID := r.PathValue("uploadID")
	key, _ := utils.DecodeBase64Key(r.PathValue("tokenKey"))
	partNum, err := strconv.Atoi(r.PathValue("partNum"))

	if err != nil {
		utils.WriteError(w, types.BadRequestError)
		return
	}
	upload, err := fb.GetMultipartUpload(key, uploadID)
	if err != nil {
		utils.WriteError(w, types.NewUnknownError(err.Error()))
		return
	}

	uploadedPart, err := upload.UploadPart(partNum, r.Body)
	if err != nil {
		utils.WriteError(w, types.NewUnknownError(err.Error()))
		return
	}

	utils.WriteJSON(w, http.StatusOK, UploadPartResponse{Etag: uploadedPart.ETag})
}

func handleCompleteUpload(db *gorm.DB, fb types.FileBucket, w http.ResponseWriter, r *http.Request) {
	uploadID := r.PathValue("uploadID")
	key, _ := utils.DecodeBase64Key(r.PathValue("tokenKey"))

	var req UploadPartRequest
	if err := utils.ReadJSON(r, &req); err != nil {
		utils.WriteError(w, types.BadRequestError)
		return
	}

	upload, err := fb.GetMultipartUpload(key, uploadID)
	if err != nil {
		utils.WriteError(w, types.NewUnknownError(err.Error()))
		return
	}

	_, err = upload.Complete(req.Parts)
	if err != nil {
		utils.WriteError(w, types.NewUnknownError(err.Error()))
		return
	}

	utils.WriteJSON(w, http.StatusOK, CompleteUploadResponse{UploadID: uploadID})
}
