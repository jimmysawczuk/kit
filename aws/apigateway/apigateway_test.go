package apigateway

import (
	"net/http"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/require"
)

var okHandler http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte(`{"success":true}`))
}

var imgHandler http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "image/jpeg")
	w.WriteHeader(200)
	w.Write(nil)
}

func buildRequest(method, path, protocol string, headers map[string]string) events.APIGatewayV2HTTPRequest {
	req := events.APIGatewayV2HTTPRequest{}
	req.RequestContext.HTTP.Method = method
	req.RequestContext.HTTP.Path = path
	req.RequestContext.HTTP.Protocol = protocol
	req.Headers = headers
	return req
}

func TestHTTPHandler(t *testing.T) {
	var tests = []struct {
		name             string
		handler          http.Handler
		req              events.APIGatewayV2HTTPRequest
		expectedResponse events.APIGatewayV2HTTPResponse
	}{
		{
			name:    "OkHandler",
			handler: okHandler,
			req:     buildRequest("GET", "/", "HTTP/1.1", map[string]string{}),
			expectedResponse: events.APIGatewayV2HTTPResponse{
				StatusCode: 200,
				Headers: map[string]string{
					"Content-Type": "application/json; charset=utf-8",
				},
				Body: `{"success":true}`,
			},
		},
		{
			name:    "ImgHandler",
			handler: imgHandler,
			req:     buildRequest("GET", "/", "HTTP/1.1", map[string]string{}),
			expectedResponse: events.APIGatewayV2HTTPResponse{
				StatusCode: 200,
				Headers: map[string]string{
					"Content-Type": "image/jpeg",
				},
				IsBase64Encoded: true,
				Body:            ``,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			f := HTTPHandler(test.handler)
			resp, err := f(test.req)
			require.NoError(t, err)
			require.Equal(t, test.expectedResponse, resp)
		})
	}
}
