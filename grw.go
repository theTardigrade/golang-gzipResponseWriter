package grw

import (
	"bytes"
	"compress/gzip"
	"net/http"
	"sync"
)

type GzipResponseWriter struct {
	gzipWriter *gzip.Writer
	http.ResponseWriter

	buffer      *bytes.Buffer
	bufferMutex sync.Mutex
}

func New(rw http.ResponseWriter) *GzipResponseWriter {
	grw := GzipResponseWriter{
		ResponseWriter: rw,
		buffer:         &bytes.Buffer{},
	}

	gw, err := gzip.NewWriterLevel(grw.buffer, gzip.BestSpeed)
	if err != nil {
		panic(err)
	}
	grw.gzipWriter = gw

	return &grw
}

func (grw *GzipResponseWriter) Close() error {
	return grw.gzipWriter.Close()
}

func (grw *GzipResponseWriter) compressBytes(b []byte) (c []byte, err error) {
	defer grw.bufferMutex.Unlock()
	grw.bufferMutex.Lock()

	grw.buffer.Reset()

	if _, err = grw.gzipWriter.Write(b); err != nil {
		return
	}

	if err = grw.gzipWriter.Flush(); err != nil {
		return
	}

	c = grw.buffer.Bytes()

	return
}

func (grw *GzipResponseWriter) Write(b []byte) (n int, err error) {
	b, err = grw.compressBytes(b)
	if err != nil {
		return
	}

	if len(b) > 0 {
		return grw.ResponseWriter.Write(b)
	}

	return
}

func (grw *GzipResponseWriter) Push(target string, opts *http.PushOptions) error {
	if pusher, ok := grw.ResponseWriter.(http.Pusher); ok {
		return pusher.Push(target, opts)
	}

	return http.ErrNotSupported
}

func (grw *GzipResponseWriter) SetHeaders() {
	h := grw.ResponseWriter.Header()

	h["Content-Encoding"] = []string{"gzip"}
	h["Vary"] = []string{"Accept-Encoding"}
}

func (grw *GzipResponseWriter) UnsetHeaders() {
	h := grw.ResponseWriter.Header()

	delete(h, "Content-Encoding")
	delete(h, "Vary")
}
