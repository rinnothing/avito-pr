package server

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"github.com/rinnothing/avito-pr/api/gen"
	"github.com/rinnothing/avito-pr/pkg/logger"
)

type Err struct {
	Code    gen.ErrorResponseErrorCode `json:"code"`
	Message string                     `json:"message"`
}

type ErrResponse struct {
	Error Err `json:"error"`
}

func newError(code gen.ErrorResponseErrorCode, message string) *ErrResponse {
	return &ErrResponse{
		Error: Err{
			Code:    code,
			Message: message,
		},
	}
}

func unauthorized(c echo.Context, reason string) error {
	return c.JSON(http.StatusUnauthorized, newError(gen.NOCANDIDATE, reason))
}

func badRequest(reason string) error {
	return echo.NewHTTPError(http.StatusBadRequest, reason)
}

func notFound(c echo.Context, reason string) error {
	return c.JSON(http.StatusNotFound, newError(gen.NOTFOUND, reason))
}

func conflict(c echo.Context, reason string, code gen.ErrorResponseErrorCode) error {
	return c.JSON(http.StatusConflict, newError(code, reason))
}

func alreadyExists(c echo.Context, reason string) error {
	return c.JSON(http.StatusBadRequest, newError(gen.TEAMEXISTS, reason))
}

func internalError() error {
	return echo.NewHTTPError(http.StatusInternalServerError)
}

func reportInternalError(ctx context.Context, err error) error {
	logger.ErrorCtx(ctx, "got internal error", zap.Error(err))
	return internalError()
}
