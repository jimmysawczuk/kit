package apigateway

import (
	"bytes"
	"context"
	"encoding/base64"
	"io"
	"net/http"
	"net/textproto"
	"net/url"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/pkg/errors"
)

// HTTPHandler shims an http.Handler into a API Gateway V2 request/response workflow.
func HTTPHandler(handler http.Handler) func(events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	return func(ar events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
		w := &responseWriter{
			statusCode: 0,
			header:     http.Header{},
			buf:        &bytes.Buffer{},
		}

		r := &http.Request{
			Method:     ar.RequestContext.HTTP.Method,
			RequestURI: ar.RequestContext.HTTP.Path,
			Proto:      ar.RequestContext.HTTP.Protocol,
			Header:     multiValueHeader(ar.Headers),
		}

		if ar.Body != "" && ar.IsBase64Encoded {
			r.Body = io.NopCloser(base64.NewDecoder(base64.StdEncoding, strings.NewReader(ar.Body)))
		} else if ar.Body != "" && !ar.IsBase64Encoded {
			r.Body = io.NopCloser(strings.NewReader(ar.Body))
		}

		var ok bool
		r.ProtoMajor, r.ProtoMinor, ok = http.ParseHTTPVersion(r.Proto)
		if !ok {
			return events.APIGatewayV2HTTPResponse{}, errors.Errorf("http: parse http version: couldn't parse version %s", r.Proto)
		}

		var err error
		r.URL, err = url.ParseRequestURI(r.RequestURI)
		if err != nil {
			return events.APIGatewayV2HTTPResponse{}, errors.Wrap(err, "url: parse request uri")
		}

		r = r.WithContext(context.Background())

		handler.ServeHTTP(w, r)

		if w.statusCode == 0 {
			w.statusCode = http.StatusOK
		}

		resp := events.APIGatewayV2HTTPResponse{
			StatusCode: w.statusCode,
			Headers:    singleValueHeader(w.header),
		}

		if w.isBinary {
			resp.Body = base64.StdEncoding.EncodeToString(w.buf.Bytes())
			resp.IsBase64Encoded = true
		} else {
			resp.Body = w.buf.String()
		}

		return resp, nil
	}
}

func multiValueHeader(in map[string]string) map[string][]string {
	tbr := make(map[string][]string, len(in))
	for k, v := range in {
		tbr[textproto.CanonicalMIMEHeaderKey(k)] = []string{v}
	}
	return tbr
}

func singleValueHeader(in map[string][]string) map[string]string {
	tbr := make(map[string]string, len(in))
	for k, v := range in {
		tbr[textproto.CanonicalMIMEHeaderKey(k)] = strings.Join(v, ", ")
	}
	return tbr
}
