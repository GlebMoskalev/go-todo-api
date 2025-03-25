package utils

import (
	"context"
	"github.com/GlebMoskalev/go-todo-api/internal/utils/contextutils"
	"log/slog"
)

func SetupLogger(ctx context.Context, baseLogger *slog.Logger, layer, operation string, extraFields ...any) *slog.Logger {
	logger := baseLogger.With("layer", layer, "operation", operation)
	if len(extraFields) > 0 {
		logger = logger.With(extraFields...)
	}
	if id := contextutils.GetRequestId(ctx); id != "" {
		logger = logger.With("request_id", id)
	}
	return logger
}
