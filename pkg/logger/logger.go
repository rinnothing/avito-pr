package logger

import (
	"context"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func ConstructLogger(level, filepath string) (*zap.Logger, error) {
	lvl, err := zapcore.ParseLevel(level)
	if err != nil {
		return nil, err
	}

	consoleEncoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())

	consoleCore := zapcore.NewCore(
		consoleEncoder,
		zapcore.Lock(os.Stdout),
		lvl,
	)

	// returning early if file is not specified
	if filepath == "" {
		return zap.New(consoleCore, zap.AddCaller()), nil
	}

	file, _ := os.OpenFile(filepath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	fileEncoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())

	fileCore := zapcore.NewCore(
		fileEncoder,
		zapcore.Lock(file),
		zapcore.DebugLevel,
	)

	core := zapcore.NewTee(consoleCore, fileCore)

	logger := zap.New(core, zap.AddCaller())

	return logger, nil
}

type key struct{}

func ErrorCtx(ctx context.Context, msg string, fields ...zap.Field) {
	FromContext(ctx).Error(msg, fields...)
}

func DebugCtx(ctx context.Context, msg string, fields ...zap.Field) {
	FromContext(ctx).Debug(msg, fields...)
}

func InfoCtx(ctx context.Context, msg string, fields ...zap.Field) {
	FromContext(ctx).Info(msg, fields...)
}

func NewContext(ctx context.Context, l *zap.Logger) context.Context {
	return context.WithValue(ctx, key{}, l)
}

func FromContext(ctx context.Context) *zap.Logger {
	if ctx == nil {
		return zap.NewNop()
	}

	v := ctx.Value(key{})
	if v == nil {
		return zap.NewNop()
	}

	return v.(*zap.Logger)
}
