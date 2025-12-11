package file

import (
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/Shua-github/Tap-Cloud-Server/core/utils"
)

func RegisterRoutes(mux *http.ServeMux, db *utils.Db, bucket string, fb utils.FileBucket) {
	db.AutoMigrate(&FileToken{})
	mux.HandleFunc("POST /1.1/fileTokens", func(w http.ResponseWriter, r *http.Request) { handleCreateFileToken(db, bucket, w, r) })
	mux.HandleFunc("DELETE /1.1/files/{ObjectID}", func(w http.ResponseWriter, r *http.Request) { handleDeleteFile(db, fb, w, r) })
	mux.HandleFunc("GET /1.1/files/{ObjectID}", func(w http.ResponseWriter, r *http.Request) { handleGetFile(fb, w, r) })

	mux.HandleFunc("POST /1.1/fileCallback", handleFileCallback)

	mux.HandleFunc("POST /buckets/{bucket}/objects/{tokenKey}/uploads", func(w http.ResponseWriter, r *http.Request) { handleStartUpload(db, fb, w, r) })
	mux.HandleFunc("PUT /buckets/{bucket}/objects/{tokenKey}/uploads/{uploadID}/{partNum}", func(w http.ResponseWriter, r *http.Request) { handleUploadPart(db, fb, w, r) })
	mux.HandleFunc("POST /buckets/{bucket}/objects/{tokenKey}/uploads/{uploadID}", func(w http.ResponseWriter, r *http.Request) { handleCompleteUpload(db, fb, w, r) })
}

func handleCreateFileToken(db *utils.Db, bucket string, w http.ResponseWriter, r *http.Request) {
	var req FileTokenRequest
	if err := utils.ReadJSON(r, &req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	sharedID := utils.RandomObjectID()

	fileToken := utils.RandomObjectID()

	var scheme string
	if r.TLS != nil {
		scheme = "https"
	} else {
		scheme = "http"
	}

	baseURL := url.URL{Scheme: scheme, Host: r.Host}

	fileURL := baseURL.String() + "/1.1/files/" + sharedID

	ft := new(FileToken)

	ft.Bucket = bucket
	ft.Key = sharedID      // Key = ObjectID
	ft.ObjectID = sharedID // ObjectID = Key
	ft.MetaData = req.MetaData
	ft.Name = req.Name
	ft.Token = fileToken
	ft.UploadURL = baseURL.String()
	ft.FileURL = fileURL
	ft.ACL = req.ACL

	if err := db.Create(&ft).Error; err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusCreated, ft)
}

func handleGetFile(fb utils.FileBucket, w http.ResponseWriter, r *http.Request) {
	objectID := r.PathValue("ObjectID")

	fileObj, err := fb.Get(objectID)
	if err != nil {
		utils.WriteError(w, http.StatusNotFound, "File not found")
		return
	}
	defer fileObj.Body.Close()

	w.Header().Set("Content-Type", "application/octet-stream")
	w.WriteHeader(http.StatusOK)

	_, err = io.Copy(w, fileObj.Body)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to send file")
		return
	}
}

func handleDeleteFile(db *utils.Db, fb utils.FileBucket, w http.ResponseWriter, r *http.Request) {
	objectID := r.PathValue("ObjectID")

	err := fb.Delete(objectID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if err := db.Where("object_id = ?", objectID).Delete(&FileToken{}).Error; err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to delete file token")
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]any{})
}

func handleFileCallback(w http.ResponseWriter, r *http.Request) {
	utils.WriteJSON(w, http.StatusOK, FileCallbackResponse{Result: true})
}

func handleStartUpload(db *utils.Db, fb utils.FileBucket, w http.ResponseWriter, r *http.Request) {
	key, _ := utils.DecodeBase64Key(r.PathValue("tokenKey"))

	var ft FileToken
	if err := db.Where("key = ?", key).First(&ft).Error; err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	_, uploadID, err := fb.CreateMultipartUpload(ft.Key)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	ft.Token = utils.RandomObjectID()
	if err := db.Save(&ft).Error; err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, StartUploadResponse{UploadID: uploadID})
}

func handleUploadPart(db *utils.Db, fb utils.FileBucket, w http.ResponseWriter, r *http.Request) {
	uploadID := r.PathValue("uploadID")
	key, _ := utils.DecodeBase64Key(r.PathValue("tokenKey"))
	partNum, err := strconv.Atoi(r.PathValue("partNum"))

	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid part number")
		return
	}
	upload, err := fb.GetMultipartUpload(key, uploadID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	uploadedPart, err := upload.UploadPart(partNum, r.Body)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, UploadPartResponse{Etag: uploadedPart.ETag})
}

func handleCompleteUpload(db *utils.Db, fb utils.FileBucket, w http.ResponseWriter, r *http.Request) {
	uploadID := r.PathValue("uploadID")
	key, _ := utils.DecodeBase64Key(r.PathValue("tokenKey"))

	var req UploadPartRequest
	if err := utils.ReadJSON(r, &req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	upload, err := fb.GetMultipartUpload(key, uploadID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	_, err = upload.Complete(req.Parts)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, CompleteUploadResponse{UploadID: uploadID})
}
