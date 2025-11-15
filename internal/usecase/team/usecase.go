package team

import (
	"context"
	"errors"

	"github.com/rinnothing/avito-pr/internal/model"
	"github.com/rinnothing/avito-pr/pkg/transaction"
)

type Usecase interface {
	CreateTeam(context.Context, *model.Team, []model.User) (*model.Team, []model.User, error)
	GetTeam(context.Context, model.TeamName) (*model.Team, []model.User, error)
}

type TeamRepo interface {
	CreateTeam(context.Context, *model.Team) error
	GetTeam(context.Context, model.TeamName) (*model.Team, error)
}

type UserRepo interface {
	CreateUser(context.Context, *model.User) error
	UpdateUser(context.Context, *model.User) error
	GetUser(context.Context, model.UserId) (*model.User, error)
}

var _ Usecase = &impl{}

type impl struct {
	teamR TeamRepo
	userR UserRepo
	tr    transaction.Transactor
}

func New(team TeamRepo, user UserRepo, tr transaction.Transactor) *impl {
	return &impl{teamR: team, userR: user, tr: tr}
}

func (u *impl) CreateTeam(ctx context.Context, team *model.Team, users []model.User) (*model.Team, []model.User, error) {
	err := u.tr.DoAtomically(ctx, func(ctx context.Context) error {
		var err error
		for _, user := range users {
			err = u.userR.CreateUser(ctx, &user)
			if errors.Is(err, model.ErrAlreadyExists) {
				err = u.userR.UpdateUser(ctx, &user)
			}
			if err != nil {
				return err
			}
		}

		// don't need to create team when it's already existing
		return u.teamR.CreateTeam(ctx, team)
	})

	return team, users, err
}

func (u *impl) GetTeam(ctx context.Context, name model.TeamName) (*model.Team, []model.User, error) {
	var team *model.Team
	var users []model.User
	err := u.tr.DoAtomically(ctx, func(ctx context.Context) error {
		var err error
		team, err = u.teamR.GetTeam(ctx, name)
		if err != nil {
			return err
		}

		users = make([]model.User, 0, len(team.Members))
		for _, id := range team.Members {
			user, err := u.userR.GetUser(ctx, id)
			if err != nil {
				return err
			}

			users = append(users, *user)
		}
		return nil
	})

	return team, users, err
}
