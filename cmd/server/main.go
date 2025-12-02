package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/Shua-github/Tap-Cloud-Server/core"
	"github.com/Shua-github/Tap-Cloud-Server/core/utils"
	"github.com/nutsdb/nutsdb"
)

type NutsDb struct {
	name string
	db   *nutsdb.DB
}

type NutsTable struct {
	bucket string
	db     *nutsdb.DB
}

func NewNutsDb(name string) utils.Db {
	options := nutsdb.DefaultOptions
	options.Dir = "./data/" + name
	options.EntryIdxMode = nutsdb.HintKeyAndRAMIdxMode
	options.SegmentSize = 256 * 1024 * 1024

	db, err := nutsdb.Open(options)
	if err != nil {
		panic(err)
	}

	return &NutsDb{name: name, db: db}
}

func (db *NutsDb) NewTable(name string) utils.Table {
	db.db.Update(func(tx *nutsdb.Tx) error {
		_ = tx.NewKVBucket(name)
		return nil
	})
	return &NutsTable{bucket: name, db: db.db}
}

func (t *NutsTable) Get(key string) ([]byte, error) {
	var value []byte

	err := t.db.View(func(tx *nutsdb.Tx) error {
		entry, err := tx.Get(t.bucket, []byte(key))
		if err != nil {
			return err
		}
		value = append([]byte{}, entry...)
		return nil
	})

	if err != nil {
		return nil, errors.New("key not found")
	}

	return value, nil
}

func (t *NutsTable) Put(key string, value []byte) error {
	return t.db.Update(func(tx *nutsdb.Tx) error {
		return tx.Put(t.bucket, []byte(key), value, 0)
	})
}

func (t *NutsTable) Del(key string) error {
	return t.db.Update(func(tx *nutsdb.Tx) error {
		return tx.Delete(t.bucket, []byte(key))
	})
}

func (t *NutsTable) Map() map[string][]byte {
	result := make(map[string][]byte)

	t.db.View(func(tx *nutsdb.Tx) error {
		entries, err := tx.PrefixScan(t.bucket, []byte(""), 0, 100)
		if err != nil {
			return nil
		}
		for k, v := range entries {
			result[fmt.Sprint(k)] = v
		}
		return nil
	})
	return result
}

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
	body       []byte
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func (lrw *loggingResponseWriter) Write(b []byte) (int, error) {
	n, err := lrw.ResponseWriter.Write(b)
	lrw.body = append(lrw.body, b...)
	return n, err
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var buf bytes.Buffer
		body := io.TeeReader(r.Body, &buf)
		data, _ := io.ReadAll(body)
		r.Body = io.NopCloser(&buf)

		lrw := &loggingResponseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(lrw, r)

		log.Printf("Request Body: %s", string(data))
		log.Printf("Response Status: %d", lrw.statusCode)
		log.Printf("Response Body: %s", string(lrw.body))
	})
}

func main() {
	bucket := os.Getenv("BUCKET")
	if bucket == "" {
		panic("BUCKET environment variable not set")
	}

	domain := os.Getenv("DOMAIN")
	if domain == "" {
		panic("DOMAIN environment variable not set")
	}

	certFile := "./" + domain + ".crt"
	keyFile := "./" + domain + ".key"

	handler := &core.Handler{
		NewDb:  NewNutsDb,
		Bucket: bucket,
	}

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	loggedMux := loggingMiddleware(mux)

	serverAddr := "0.0.0.0:443"
	log.Printf("Server running at https://%s\n", serverAddr)
	if err := http.ListenAndServeTLS(serverAddr, certFile, keyFile, loggedMux); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
