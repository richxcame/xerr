package xerr

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

type APIError interface {
	error
	StatusCode() int
	ErrorKey() string
}

type HttpError struct {
	Code     int            `json:"-"`
	Key      string         `json:"id"`
	Message  string         `json:"message"`
	Lang     string         `json:"-"`
	Internal error          `json:"-"`
	Meta     map[string]any `json:"meta,omitempty"`
	TraceID  string         `json:"trace_id,omitempty"`
	UserID   string         `json:"user_id,omitempty"`
	Group    string         `json:"group,omitempty"`
}

func (e *HttpError) Error() string {
	return fmt.Sprintf("[%d] %s: %s", e.Code, e.Key, e.Message)
}

func (e *HttpError) Unwrap() error {
	return e.Internal
}

func (e *HttpError) StatusCode() int {
	return e.Code
}

func (e *HttpError) ErrorKey() string {
	return e.Key
}

func (e *HttpError) Localized() string {
	msg := i18nLookup(e.Key, e.Lang)
	if msg == "" {
		return e.Message
	}
	return msg
}

func (e *HttpError) AddMeta(k string, v interface{}) {
	if e.Meta == nil {
		e.Meta = make(map[string]interface{})
	}
	e.Meta[k] = v
}

func (e *HttpError) GetMeta(k string) (interface{}, bool) {
	if e.Meta == nil {
		return nil, false
	}
	v, ok := e.Meta[k]
	return v, ok
}

func (e *HttpError) ToJSON() string {
	b, _ := json.Marshal(e)
	return string(b)
}

func New(code int, key, msg, lang string, internal error, meta map[string]interface{}) *HttpError {
	return &HttpError{
		Code:     code,
		Key:      key,
		Message:  msg,
		Lang:     lang,
		Internal: internal,
		Meta:     meta,
	}
}

func BadRequest(key, msg, lang string, meta map[string]interface{}) *HttpError {
	return New(http.StatusBadRequest, key, msg, lang, nil, meta)
}

func ValidationError(errors map[string]string) *HttpError {
	return New(http.StatusBadRequest, "validation_error", "validation failed", DefaultLang, nil, map[string]any{"fields": errors})
}

func NotFound(key, msg, lang string) *HttpError {
	return New(http.StatusNotFound, key, msg, lang, nil, nil)
}

func Internal(err error, meta map[string]interface{}) *HttpError {
	return New(http.StatusInternalServerError, "internal_error", "internal server error", "en", fmt.Errorf("%w", err), meta)
}

func Unauthorized(msg string) *HttpError {
	return New(http.StatusUnauthorized, "unauthorized", msg, "en", nil, nil)
}

func Forbidden(msg string) *HttpError {
	return New(http.StatusForbidden, "forbidden", msg, "en", nil, nil)
}

func WithTrace(err *HttpError, traceID, userID string) *HttpError {
	err.TraceID = traceID
	err.UserID = userID
	return err
}

func Is(err error, key string) bool {
	var e *HttpError
	return errors.As(err, &e) && e.Key == key
}

func AsHttpError(err error) (*HttpError, bool) {
	var httpErr *HttpError
	if errors.As(err, &httpErr) {
		return httpErr, true
	}
	return nil, false
}
