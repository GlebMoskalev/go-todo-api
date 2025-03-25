package contextutils

import "context"

const requestIDContextKey = "request_id"

func GetRequestId(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if requestId, ok := ctx.Value(requestIDContextKey).(string); ok {
		return requestId
	}
	return ""
}

func WithRequestId(ctx context.Context, requestId string) context.Context {
	return context.WithValue(ctx, requestIDContextKey, requestId)
}
