package file

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"

	"github.com/Shua-github/Tap-Cloud-Server/core/general"
	"github.com/Shua-github/Tap-Cloud-Server/core/utils"
)

func RegisterRoutes(mux *http.ServeMux, db utils.Db, bucket string) {
	mux.HandleFunc("POST /1.1/fileTokens", func(w http.ResponseWriter, r *http.Request) { handleCreateFileToken(db, bucket, w, r) })
	mux.HandleFunc("DELETE /1.1/files/{ObjectID}", func(w http.ResponseWriter, r *http.Request) { handleDeleteFile(db, w, r) })
	mux.HandleFunc("GET /1.1/files/{ObjectID}", func(w http.ResponseWriter, r *http.Request) { handleGetFile(db, w, r) })

	mux.HandleFunc("POST /1.1/fileCallback", handleFileCallback)

	mux.HandleFunc("POST /buckets/{bucket}/objects/{tokenKey}/uploads", func(w http.ResponseWriter, r *http.Request) { handleStartUpload(db, w, r) })
	mux.HandleFunc("PUT /buckets/{bucket}/objects/{tokenKey}/uploads/{uploadID}/{partNum}", func(w http.ResponseWriter, r *http.Request) { handleUploadPart(db, w, r) })
	mux.HandleFunc("POST /buckets/{bucket}/objects/{tokenKey}/uploads/{uploadID}", func(w http.ResponseWriter, r *http.Request) { handleCompleteUpload(db, w, r) })
}

func handleCreateFileToken(db utils.Db, bucket string, w http.ResponseWriter, r *http.Request) {
	var req FileTokenRequest
	if err := utils.ReadJSON(r, &req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	fileToken := utils.RandomObjectID()
	fileTokenKey := utils.RandomObjectID()
	fileObjectID := utils.RandomObjectID()

	var scheme string
	if r.TLS != nil {
		scheme = "https"
	} else {
		scheme = "http"
	}

	baseURL := url.URL{Scheme: scheme, Host: r.Host}

	fileURL := baseURL.String() + "/1.1/files/" + fileObjectID

	now := utils.GetUTCISO()

	ft := new(general.File)

	ft.Bucket = bucket
	ft.CreatedAt = now
	ft.Key = fileTokenKey
	ft.MetaData = req.MetaData
	ft.MimeType = "application/octet-stream"
	ft.Name = req.Name
	ft.ObjectID = fileObjectID
	ft.Provider = "qiniu"
	ft.Token = fileToken
	ft.UploadURL = baseURL.String()
	ft.FileURL = fileURL

	ftm := utils.Bind(db.NewTable("filetoken"), fileTokenKey, ft)
	if err := ftm.Save(); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "db error")
		return
	}

	utils.WriteJSON(w, http.StatusCreated, ft)
}

func handleGetFile(db utils.Db, w http.ResponseWriter, r *http.Request) {
	ObjectID := r.PathValue("ObjectID")

	file := new(File)
	fm := utils.Bind(db.NewTable("file"), ObjectID, file)

	if err := fm.Load(); err != nil {
		utils.WriteError(w, http.StatusNotFound, "file not found")
		return
	}

	if !file.OK {
		utils.WriteError(w, http.StatusNotFound, "file not completed")
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.WriteHeader(http.StatusOK)
	w.Write(file.Data)
}

func handleDeleteFile(db utils.Db, w http.ResponseWriter, r *http.Request) {
	ObjectID := r.PathValue("ObjectID")

	fm := utils.Bind(db.NewTable("file"), ObjectID, new(File))

	if err := fm.Load(); err != nil {
		utils.WriteError(w, http.StatusNotFound, "file not found")
		return
	}

	_ = db.NewTable("filetoken").Del(fm.V.Key)
	_ = fm.Delete()

	utils.WriteJSON(w, http.StatusOK, map[any]any{})
}

func handleFileCallback(w http.ResponseWriter, r *http.Request) {
	utils.WriteJSON(w, http.StatusOK, FileCallbackResponse{Result: true})
}

func handleStartUpload(db utils.Db, w http.ResponseWriter, r *http.Request) {
	key, _ := utils.DecodeBase64Key(r.PathValue("tokenKey"))
	ftm := utils.Bind(db.NewTable("filetoken"), key, new(general.File))

	if err := ftm.Load(); err != nil {
		utils.WriteError(w, http.StatusNotFound, fmt.Errorf("%v,err:%w", db.NewTable("filetoken").Map(), err).Error())
	}

	uploadID := utils.RandomObjectID()
	usm := utils.Bind(db.NewTable("upload"), uploadID, new(UploadSession))
	usm.V.UploadID = uploadID
	if err := usm.Save(); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, StartUploadResponse{UploadID: uploadID})
}

func handleUploadPart(db utils.Db, w http.ResponseWriter, r *http.Request) {
	uploadID := r.PathValue("uploadID")
	key, _ := utils.DecodeBase64Key(r.PathValue("tokenKey"))
	ftm := utils.Bind(db.NewTable("filetoken"), key, new(general.File))
	if err := ftm.Load(); err != nil {
		utils.WriteError(w, http.StatusNotFound, err.Error())
		return
	}
	ObjectID := ftm.V.ObjectID
	partData, err := io.ReadAll(r.Body)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "failed to read part data")
		return
	}

	file := new(File)
	fm := utils.Bind(db.NewTable("file"), ObjectID, file)

	if err := fm.Load(); err != nil {
		file.Data = partData
		file.UploadID = uploadID
		file.PartsReceived = 1
		file.OK = false
	} else {
		file.Data = append(file.Data, partData...)
		file.PartsReceived++
	}

	if err := fm.Save(); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	etag := utils.RandomObjectID()
	utils.WriteJSON(w, http.StatusOK, UploadPartResponse{Etag: etag})
}

func handleCompleteUpload(db utils.Db, w http.ResponseWriter, r *http.Request) {
	uploadID := r.PathValue("uploadID")
	key, _ := utils.DecodeBase64Key(r.PathValue("tokenKey"))
	ftm := utils.Bind(db.NewTable("filetoken"), key, new(general.File))
	if err := ftm.Load(); err != nil {
		utils.WriteError(w, http.StatusNotFound, "file token not found")
		return
	}

	ObjectID := ftm.V.ObjectID

	usm := utils.Bind(db.NewTable("upload"), uploadID, new(UploadSession))
	if err := usm.Load(); err != nil {
		utils.WriteError(w, http.StatusNotFound, "upload session not found")
		return
	} else {
		if err := usm.Delete(); err != nil {
			log.Println("failed to delete upload session:", err)
		}
	}

	fm := utils.Bind(db.NewTable("file"), ObjectID, new(File))
	if err := fm.Load(); err != nil {
		utils.WriteError(w, http.StatusNotFound, "file not found")
		return
	}

	fm.V.OK = true
	fm.V.Key = ftm.V.Key
	if err := fm.Save(); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "db error")
		return
	}

	utils.WriteJSON(w, http.StatusOK, CompleteUploadResponse{UploadID: uploadID})
}
