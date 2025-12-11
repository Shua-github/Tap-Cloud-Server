package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
)

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
		logBuf.WriteString("Path: " + r.URL.Path + "\n")
		logBuf.WriteString("Method: " + r.Method + "\n")
		logBuf.WriteString("RemoteAddr: " + r.RemoteAddr + "\n")
		logBuf.WriteString("UserAgent: " + r.UserAgent() + "\n")
		if len(reqData) <= 1024 {
			logBuf.WriteString("Request Body: " + string(reqData) + "\n")
		} else {
			logBuf.WriteString("Request Body: [TOO LARGE]" + "\n")
		}
		logBuf.WriteString("Response Status: ")
		logBuf.WriteString(http.StatusText(lrw.statusCode))
		logBuf.WriteString(" (" +
			func(code int) string { return fmt.Sprintf("%d", code) }(lrw.statusCode) +
			")\n")
		if len(lrw.body) <= 1024 {
			logBuf.WriteString("Response Body: " + string(lrw.body) + "\n")
		} else {
			logBuf.WriteString("Response Body: [TOO LARGE]\n")
		}
		log.Print(logBuf.String())
	})
}
