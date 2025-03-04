package web_test

import (
	"bytes"
	"io"
)

func getBody(r io.ReadCloser) string {
	buf := bytes.Buffer{}
	io.Copy(&buf, r)
	r.Close()
	return buf.String()
}
