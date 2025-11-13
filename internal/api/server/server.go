package server

import (
	"context"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"github.com/rinnothing/avito-pr/api/gen"
	"github.com/rinnothing/avito-pr/internal/model"
	"github.com/rinnothing/avito-pr/internal/usecase/pullrequest"
	"github.com/rinnothing/avito-pr/internal/usecase/team"
	"github.com/rinnothing/avito-pr/internal/usecase/user"
	"github.com/rinnothing/avito-pr/pkg/logger"
)

var _ gen.ServerInterface = &serverImplementation{}

type serverImplementation struct {
	l *zap.Logger

	user user.Usecase
	team team.Usecase
	pr   pullrequest.Usecase
}

func New(user user.Usecase, team team.Usecase, pr pullrequest.Usecase, l *zap.Logger) *serverImplementation {
	return &serverImplementation{l: l, user: user, team: team, pr: pr}
}

func (s *serverImplementation) withLogger(ctx context.Context) context.Context {
	return logger.NewContext(ctx, s.l)
}

func isAdmin(ctx echo.Context) bool {
	return ctx.Get(gen.AdminTokenScopes) != nil
}

func isUser(ctx echo.Context) bool {
	return ctx.Get(gen.UserTokenScopes) != nil
}

func isAdminOrUser(ctx echo.Context) bool {
	return isUser(ctx) || isAdmin(ctx)
}

func mapShortStatus(st model.PRStatus) gen.PullRequestShortStatus {
	switch st {
	case model.PRMerged:
		return gen.PullRequestShortStatusMERGED
	case model.PROpen:
		return gen.PullRequestShortStatusOPEN
	}
	return "WRONG_STATUS"
}

func mapStatus(st model.PRStatus) gen.PullRequestStatus {
	switch st {
	case model.PRMerged:
		return gen.PullRequestStatusMERGED
	case model.PROpen:
		return gen.PullRequestStatusOPEN
	}
	return "WRONG_STATUS"
}
