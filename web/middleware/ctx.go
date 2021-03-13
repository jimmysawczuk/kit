package middleware

type ctxKey int

const (
	// RequestIDKey is the context key for a given request ID.
	RequestIDKey ctxKey = iota

	// StartTimeKey is a context key for storing when the request starts.
	StartTimeKey
)
