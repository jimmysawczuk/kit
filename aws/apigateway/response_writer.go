package apigateway

import (
	"bytes"
	"net/http"
	"strings"
)

type responseWriter struct {
	statusCode  int
	header      http.Header
	buf         *bytes.Buffer
	contentType string
	isBinary    bool
}

// Make sure our responseWriter is implementing http.ResponseWriter.
var _ http.ResponseWriter = &responseWriter{}

func (wr *responseWriter) Header() http.Header {
	return wr.header
}

func (wr *responseWriter) Write(by []byte) (int, error) {
	return wr.buf.Write(by)
}

func (wr *responseWriter) WriteHeader(code int) {
	wr.contentType = wr.header.Get("Content-Type")
	if strings.HasPrefix(wr.contentType, "image/") {
		wr.isBinary = true
	}

	wr.statusCode = code
}
