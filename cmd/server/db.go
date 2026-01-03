package main

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/Shua-github/Tap-Cloud-Server/core/types"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func mustNewDb(name string) (db *gorm.DB) {
	dir := "./data/db"
	if err := os.MkdirAll(dir, 0755); err != nil {
		panic(err)
	}
	path := filepath.Join(dir, name+".db")
	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	return
}

type LocalFileBucket struct {
	rootDir string
	mu      sync.RWMutex
}

func NewLocalFileBucket(name string) types.FileBucket {
	rootDir := filepath.Join("./data/file_buckets", name)
	if err := os.MkdirAll(rootDir, 0755); err != nil {
		panic(fmt.Sprintf("failed to create bucket directory: %v", err))
	}
	tmpDir := filepath.Join(rootDir, "_multipart")
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		panic(fmt.Sprintf("failed to create multipart temp directory: %v", err))
	}
	return &LocalFileBucket{
		rootDir: rootDir,
	}
}

func (b *LocalFileBucket) Get(key string) (*types.FileObject, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	dataPath := filepath.Join(b.rootDir, key+".data")
	metaPath := filepath.Join(b.rootDir, key+".meta")

	if _, err := os.Stat(dataPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("file not found: %s", key)
	}
	if _, err := os.Stat(metaPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("meta not found: %s", key)
	}

	data, err := os.ReadFile(dataPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read data file: %v", err)
	}

	metaData, err := os.ReadFile(metaPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read meta file: %v", err)
	}
	metaLines := strings.Split(string(metaData), "\n")
	var etag string
	for _, line := range metaLines {
		if strings.HasPrefix(line, "etag=") {
			etag = strings.TrimPrefix(line, "etag=")
			break
		}
	}
	if etag == "" {
		return nil, fmt.Errorf("etag not found in meta: %s", key)
	}

	return &types.FileObject{
		Key:  key,
		Body: io.NopCloser(bytes.NewReader(data)),
		ETag: etag,
	}, nil
}

func (b *LocalFileBucket) Delete(key string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	dataPath := filepath.Join(b.rootDir, key+".data")
	metaPath := filepath.Join(b.rootDir, key+".meta")

	if _, err := os.Stat(dataPath); os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s", key)
	}

	if err := os.Remove(dataPath); err != nil {
		return fmt.Errorf("failed to delete data file: %v", err)
	}
	if err := os.Remove(metaPath); err != nil {
		return fmt.Errorf("failed to delete meta file: %v", err)
	}
	return nil
}

func (b *LocalFileBucket) CreateMultipartUpload(key string) (uploadID string, err error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	uploadID = fmt.Sprintf("%d_%s", b.getCurrentTimestamp(), strings.Replace(key, "/", "_", -1))
	uploadDir := filepath.Join(b.rootDir, "_multipart", uploadID)
	if err = os.MkdirAll(uploadDir, 0755); err != nil {
		err = fmt.Errorf("failed to create upload directory: %v", err)
		return
	}
	metaFile := filepath.Join(uploadDir, "_meta.txt")
	if err = os.WriteFile(metaFile, []byte(key), 0644); err != nil {
		os.RemoveAll(uploadDir)
		err = fmt.Errorf("failed to save upload metadata: %v", err)
		return
	}
	return
}

func (b *LocalFileBucket) GetMultipartUpload(key string, uploadID string) (types.MultipartUpload, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	uploadDir := filepath.Join(b.rootDir, "_multipart", uploadID)
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("multipart upload not found: %s", uploadID)
	}
	metaFile := filepath.Join(uploadDir, "_meta.txt")
	metaData, err := os.ReadFile(metaFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read upload metadata: %v", err)
	}
	if string(metaData) != key {
		return nil, fmt.Errorf("upload key mismatch: expected %s, got %s", key, string(metaData))
	}
	return &LocalMultipartUpload{
		uploadID:  uploadID,
		key:       key,
		uploadDir: uploadDir,
		bucket:    b,
	}, nil
}

func (b *LocalFileBucket) getCurrentTimestamp() int64 {
	return time.Now().UnixNano()
}

type LocalMultipartUpload struct {
	uploadID  string
	key       string
	uploadDir string
	bucket    *LocalFileBucket
}

func (u *LocalMultipartUpload) UploadPart(partNumber int, data io.Reader) (types.UploadedPart, error) {
	u.bucket.mu.Lock()
	defer u.bucket.mu.Unlock()
	partDataFile := filepath.Join(u.uploadDir, fmt.Sprintf("part_%d.data", partNumber))
	partMetaFile := filepath.Join(u.uploadDir, fmt.Sprintf("part_%d.meta", partNumber))

	file, err := os.Create(partDataFile)
	if err != nil {
		return types.UploadedPart{}, fmt.Errorf("failed to create part data file: %v", err)
	}
	defer file.Close()

	var buf bytes.Buffer
	teeReader := io.TeeReader(data, &buf)
	hash := md5.New()
	if _, err := io.Copy(io.MultiWriter(file, hash), teeReader); err != nil {
		return types.UploadedPart{}, fmt.Errorf("failed to write part data: %v", err)
	}
	etag := fmt.Sprintf("%x", hash.Sum(nil))

	metaData := fmt.Sprintf("etag=%s\n", etag)
	if err := os.WriteFile(partMetaFile, []byte(metaData), 0644); err != nil {
		return types.UploadedPart{}, fmt.Errorf("failed to save part metadata: %v", err)
	}

	return types.UploadedPart{
		PartNumber: partNumber,
		ETag:       etag,
	}, nil
}

func (u *LocalMultipartUpload) Complete(parts []types.UploadedPart) (types.FileObject, error) {
	u.bucket.mu.Lock()
	defer u.bucket.mu.Unlock()
	sort.Slice(parts, func(i, j int) bool {
		return parts[i].PartNumber < parts[j].PartNumber
	})

	finalDataPath := filepath.Join(u.bucket.rootDir, u.key+".data")
	finalMetaPath := filepath.Join(u.bucket.rootDir, u.key+".meta")

	if err := os.MkdirAll(filepath.Dir(finalDataPath), 0755); err != nil {
		return types.FileObject{}, fmt.Errorf("failed to create final file directory: %v", err)
	}

	finalFile, err := os.Create(finalDataPath)
	if err != nil {
		return types.FileObject{}, fmt.Errorf("failed to create final data file: %v", err)
	}
	defer finalFile.Close()

	combinedHash := md5.New()
	multiWriter := io.MultiWriter(finalFile, combinedHash)

	for _, part := range parts {
		partDataFile := filepath.Join(u.uploadDir, fmt.Sprintf("part_%d.data", part.PartNumber))
		partMetaFile := filepath.Join(u.uploadDir, fmt.Sprintf("part_%d.meta", part.PartNumber))

		metaData, err := os.ReadFile(partMetaFile)
		if err != nil {
			finalFile.Close()
			os.Remove(finalDataPath)
			return types.FileObject{}, fmt.Errorf("failed to read part metadata: %v", err)
		}
		metaLines := strings.Split(string(metaData), "\n")
		var partEtag string
		for _, line := range metaLines {
			if strings.HasPrefix(line, "etag=") {
				partEtag = strings.TrimPrefix(line, "etag=")
				break
			}
		}
		if partEtag != part.ETag {
			finalFile.Close()
			os.Remove(finalDataPath)
			return types.FileObject{}, fmt.Errorf("part ETag mismatch for part %d", part.PartNumber)
		}

		partData, err := os.Open(partDataFile)
		if err != nil {
			finalFile.Close()
			os.Remove(finalDataPath)
			return types.FileObject{}, fmt.Errorf("failed to open part data file: %v", err)
		}
		if _, err := io.Copy(multiWriter, partData); err != nil {
			partData.Close()
			finalFile.Close()
			os.Remove(finalDataPath)
			return types.FileObject{}, fmt.Errorf("failed to copy part data: %v", err)
		}
		partData.Close()
	}

	finalETag := fmt.Sprintf("%x", combinedHash.Sum(nil))
	finalMetaData := fmt.Sprintf("etag=%s\n", finalETag)
	if err := os.WriteFile(finalMetaPath, []byte(finalMetaData), 0644); err != nil {
		finalFile.Close()
		os.Remove(finalDataPath)
		return types.FileObject{}, fmt.Errorf("failed to write final meta: %v", err)
	}

	os.RemoveAll(u.uploadDir)

	finalData, err := os.ReadFile(finalDataPath)
	if err != nil {
		return types.FileObject{}, fmt.Errorf("failed to read final data: %v", err)
	}

	return types.FileObject{
		Key:  u.key,
		Body: io.NopCloser(bytes.NewReader(finalData)),
		ETag: finalETag,
	}, nil
}

func (u *LocalMultipartUpload) Abort() error {
	u.bucket.mu.Lock()
	defer u.bucket.mu.Unlock()
	if err := os.RemoveAll(u.uploadDir); err != nil {
		return fmt.Errorf("failed to abort multipart upload: %v", err)
	}
	return nil
}
