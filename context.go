package xerr

import "context"

type contextKey string

const (
	ContextKeyTraceID contextKey = "trace_id"
	ContextKeyUserID  contextKey = "user_id"
)

func GetTraceID(ctx context.Context) string {
	if traceID, ok := ctx.Value(ContextKeyTraceID).(string); ok {
		return traceID
	}
	return ""
}

func GetUserID(ctx context.Context) string {
	if userID, ok := ctx.Value(ContextKeyUserID).(string); ok {
		return userID
	}
	return ""
}
