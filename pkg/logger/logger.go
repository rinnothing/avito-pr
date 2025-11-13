package logger

import (
	"context"

	"go.uber.org/zap"
)

func ErrorCtx(ctx context.Context, msg string, fields ...zap.Field) {
	panic("not implemented")
}
