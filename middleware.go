package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
)

func PanicRecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("panic recovered: %v\n%s", rec, debug.Stack())
				err := fmt.Errorf("panic recovered: %v", rec)
				WriteAPIResponse(w, nil, Internal(err, nil), "en")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func TraceContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		traceID := r.Header.Get(HeaderTraceID)
		userID := r.Header.Get(HeaderUserID)

		ctx := context.WithValue(r.Context(), ContextKeyTraceID, traceID)
		ctx = context.WithValue(ctx, ContextKeyUserID, userID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
