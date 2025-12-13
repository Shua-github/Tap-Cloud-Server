package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"unicode/utf8"
)

func tryDecode(b []byte) string {
	if utf8.Valid(b) {
		return string(b)
	} else {
		return ""
	}
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

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var reqBuf bytes.Buffer
		body := io.TeeReader(r.Body, &reqBuf)
		reqData, _ := io.ReadAll(body)
		r.Body = io.NopCloser(&reqBuf)
		lrw := &loggingResponseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}
		next.ServeHTTP(lrw, r)
		var logBuf bytes.Buffer
		logBuf.WriteString("Path: " + r.URL.RequestURI() + "\n")
		logBuf.WriteString("Method: " + r.Method + "\n")
		logBuf.WriteString("RemoteAddr: " + r.RemoteAddr + "\n")
		logBuf.WriteString("UserAgent: " + r.UserAgent() + "\n")
		logBuf.WriteString("Request Body: " + tryDecode(reqData) + "\n")
		logBuf.WriteString("Response Status: ")
		logBuf.WriteString(http.StatusText(lrw.statusCode))
		logBuf.WriteString(" (" +
			func(code int) string { return fmt.Sprintf("%d", code) }(lrw.statusCode) +
			")\n")
		logBuf.WriteString("Response Body: " + tryDecode(lrw.body) + "\n")
		log.Print(logBuf.String())
	})
}
