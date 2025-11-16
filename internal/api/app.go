package api

import (
	"context"
	"errors"
	"net"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"

	"github.com/rinnothing/avito-pr/api/gen"
	"github.com/rinnothing/avito-pr/config"
	"github.com/rinnothing/avito-pr/db"
	"github.com/rinnothing/avito-pr/internal/api/server"
	dbRepo "github.com/rinnothing/avito-pr/internal/repository/db"
	"github.com/rinnothing/avito-pr/internal/usecase/pullrequest"
	"github.com/rinnothing/avito-pr/internal/usecase/team"
	"github.com/rinnothing/avito-pr/internal/usecase/user"
	"github.com/rinnothing/avito-pr/pkg/logger"
	"github.com/rinnothing/avito-pr/pkg/transaction"
)

type Server struct {
	cancel context.CancelFunc
}

func (s *Server) Run(lg *zap.Logger, cfg *config.Config) {
	sigCtx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	ctx := logger.NewContext(sigCtx, lg)

	dbPool, err := pgxpool.New(ctx, cfg.URL)
	if err != nil {
		lg.Error("can not create pgxpool", zap.Error(err))
		return
	}
	defer dbPool.Close()

	db.SetupPostgres(dbPool, lg)

	repo := dbRepo.NewPostgresRepository(dbPool)
	transactor := transaction.NewTransactor(dbPool)

	prUsecase := pullrequest.New(repo, repo, repo, transactor)
	teamUsecase := team.New(repo, repo, transactor)
	userUsecase := user.New(repo, transactor)

	srv := server.New(userUsecase, teamUsecase, prUsecase, lg)

	e := echo.New()
	e.Use(middleware.RecoverWithConfig(middleware.RecoverConfig{
		LogErrorFunc: func(c echo.Context, err error, stack []byte) error {
			lg.Error("failed with panic", zap.String("path", c.Path()), zap.Error(err), zap.ByteString("stack", stack))
			return err
		},
	}))
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			lg.Debug("incoming request", zap.String("uri", v.URI), zap.Any("values", v.FormValues))
			return nil
		},
	}))

	e.IPExtractor = echo.ExtractIPDirect()
	gen.RegisterHandlers(e, srv)

	go func() {
		if err := e.Start(net.JoinHostPort("0.0.0.0", cfg.HTTP.Port)); !errors.Is(err, http.ErrServerClosed) {
			lg.Fatal("server died", zap.Error(err))
		}
	}()

	<-ctx.Done()

	stopCtx, stopCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer stopCancel()

	if err := e.Shutdown(stopCtx); err != nil {
		lg.Fatal("server shutdown failed", zap.Error(err))
		return
	}

	lg.Info("server shutdown")
}

func (s *Server) Stop() {
	s.cancel()
}
