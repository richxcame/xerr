package xerr

// Package errs provides a scalable, structured, multi-language-ready error handling system.
var (
	HeaderTraceID      = "X-Trace-ID"
	HeaderUserID       = "X-User-ID"
	RelaxedKeyFormat   = true
	AllowDuplicateKeys = false
	DefaultLang        = "en"
)

// ObserveError is an optional hook to collect error metrics, e.g., for Prometheus
var ObserveError func(key string, code int)
