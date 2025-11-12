package team

import (
	"context"

	"github.com/rinnothing/avito-pr/internal/model"
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
