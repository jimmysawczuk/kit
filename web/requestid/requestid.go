package requestid

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"sync/atomic"
)

type ctxKey int

const (
	requestIDKey ctxKey = 1 << iota

	defaultLength int = 8
)

// RandomPrefix returns a random prefix, a base64-encoded string of length bytes.
func RandomPrefix(length int) string {
	buf := make([]byte, length)
	rand.Read(buf)

	return fmt.Sprintf("%s", base64.RawURLEncoding.EncodeToString(buf))
}

// HostnamePrefix returns a prefix based on the system's hostname, in the following format:
//
// <hostname>/<random string>
//
// The random string is length bytes, base64-encoded.
func HostnamePrefix(length int) string {
	hostname, err := os.Hostname()
	if hostname == "" || err != nil {
		hostname = "localhost"
	}

	buf := make([]byte, length)
	rand.Read(buf)

	return fmt.Sprintf("%s/%s", hostname, base64.RawURLEncoding.EncodeToString(buf))
}

// Generator generates request IDs with the set prefix.
type Generator struct {
	Prefix string

	id uint64
}

// DefaultGenerator is the default ID generator, using the hostname and some random bytes as
// its prefix.
var DefaultGenerator = &Generator{
	Prefix: HostnamePrefix(8),
}

// Next checks the passed request for an incoming request ID; if its set, that's the request ID we use.
// Otherwise, we'll generate the next request ID and return it.
func (c *Generator) Next(r *http.Request) string {
	requestID := r.Header.Get("X-Request-Id")
	if requestID == "" {
		id := atomic.AddUint64(&c.id, 1)
		requestID = fmt.Sprintf("%s-%09d", c.Prefix, id)
	}
	return requestID
}

// Get attempts to get the current request ID from the provided context.Context. It'll return an
// empty string if the context doesn't contain a request ID.
func (c *Generator) Get(ctx context.Context) string {
	v, _ := ctx.Value(requestIDKey).(string)
	return v
}

// Set returns a copy of the provided context.Context with the provided request ID as a value.
func (c *Generator) Set(parent context.Context, id string) context.Context {
	ctx := context.WithValue(parent, requestIDKey, id)
	return ctx
}

// Next wraps DefaultGenerator.Next.
func Next(r *http.Request) string {
	return DefaultGenerator.Next(r)
}

// Get wraps DefaultGenerator.Get.
func Get(ctx context.Context) string {
	return DefaultGenerator.Get(ctx)
}

// Set wraps DefaultGenerator.Set.
func Set(parent context.Context, id string) context.Context {
	return DefaultGenerator.Set(parent, id)
}
