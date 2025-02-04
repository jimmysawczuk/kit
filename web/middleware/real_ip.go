package middleware

// This file is largely ported from https://raw.githubusercontent.com/go-chi/chi/refs/tags/v5.2.1/middleware/realip.go and adapted to fit the web.Handler signature.

import (
	"context"
	"net"
	"net/http"
	"strings"

	"github.com/jimmysawczuk/kit/web"
	"github.com/sirupsen/logrus"
)

var (
	trueClientIP  = http.CanonicalHeaderKey("True-Client-IP")
	xForwardedFor = http.CanonicalHeaderKey("X-Forwarded-For")
	xRealIP       = http.CanonicalHeaderKey("X-Real-IP")
)

// RealIP is a middleware that sets a http.Request's RemoteAddr to the results
// of parsing either the True-Client-IP, X-Real-IP or the X-Forwarded-For headers
// (in that order).
//
// This middleware should be inserted fairly early in the middleware stack to
// ensure that subsequent layers (e.g., request loggers) which examine the
// RemoteAddr will see the intended value.
//
// You should only use this middleware if you can trust the headers passed to
// you (in particular, the three headers this middleware uses), for example
// because you have placed a reverse proxy like HAProxy or nginx in front of
// chi. If your reverse proxies are configured to pass along arbitrary header
// values from the client, or if you use this middleware without a reverse
// proxy, malicious clients will be able to make you very sad (or, depending on
// how you're using RemoteAddr, vulnerable to an attack of some sort).
func RealIP(h web.Handler) web.Handler {
	return func(ctx context.Context, log logrus.FieldLogger, w http.ResponseWriter, r *http.Request) {
		if rip := realIP(r); rip != "" {
			r.RemoteAddr = rip
			ctx = context.WithValue(ctx, ipAddressKey, rip)
		}

		h(ctx, log, w, r.WithContext(ctx))
	}
}

func realIP(r *http.Request) string {
	var ip string

	if tcip := r.Header.Get(trueClientIP); tcip != "" {
		ip = tcip
	} else if xrip := r.Header.Get(xRealIP); xrip != "" {
		ip = xrip
	} else if xff := r.Header.Get(xForwardedFor); xff != "" {
		i := strings.Index(xff, ",")
		if i == -1 {
			i = len(xff)
		}
		ip = xff[:i]
	}
	if ip == "" || net.ParseIP(ip) == nil {
		return ""
	}
	return ip
}
