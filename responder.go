package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
)

func WriteAPIResponse(w http.ResponseWriter, data interface{}, err error, lang string) {
	w.Header().Set("Content-Type", "application/json")

	if err == nil {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data":    data,
		})
		return
	}

	httpErr := convertToHttpError(err, lang)

	registryMu.RLock()
	def := Registry[httpErr.Key]
	registryMu.RUnlock()

	expose := def.Expose || os.Getenv("APP_ENV") == "dev"
	key := httpErr.Key
	msg := httpErr.Localized()
	if !expose {
		key = "internal_error"
		msg = "unexpected server error"
	}

	w.WriteHeader(httpErr.Code)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success": false,
		"error": map[string]interface{}{
			"id":       key,
			"message":  msg,
			"meta":     httpErr.Meta,
			"trace_id": httpErr.TraceID,
			"user_id":  httpErr.UserID,
		},
	})

	logger.Errorf("HTTP error response: key=%s, message=%s, trace_id=%s, user_id=%s, meta=%v, err=%v",
		httpErr.Key, httpErr.Message, httpErr.TraceID, httpErr.UserID, httpErr.Meta, httpErr.Internal)

	if ObserveError != nil {
		ObserveError(httpErr.Key, httpErr.Code)
	}
}

func convertToHttpError(err error, lang string) *HttpError {
	var httpErr *HttpError
	if errors.As(err, &httpErr) {
		httpErr.Lang = lang
		return httpErr
	}
	return &HttpError{
		Code:     http.StatusInternalServerError,
		Key:      "unknown_error",
		Message:  "unexpected server error",
		Lang:     lang,
		Internal: err,
	}
}
