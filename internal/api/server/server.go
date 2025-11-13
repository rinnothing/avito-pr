package server

import (
	"context"

	"go.uber.org/zap"

	"github.com/rinnothing/avito-pr/api/gen"
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
