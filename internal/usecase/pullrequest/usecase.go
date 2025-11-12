package pullrequest

import (
	"context"

	"github.com/rinnothing/avito-pr/internal/model"
)

type Usecase interface {
	CreatePR(context.Context, *model.PullRequest) (*model.PullRequest, error)
	MergePR(context.Context, model.PRId) (*model.PullRequest, error)
	Reassign(context.Context, model.PRId, model.UserId) (*model.PullRequest, model.UserId, error)
	ListPR(context.Context, model.UserId) (*model.UserId, []model.PullRequest, error)
}

type PRRepo interface {
	CreatePR(context.Context, *model.PullRequest) error
	UpdatePR(context.Context, *model.PullRequest) error
	GetPR(context.Context, model.PRId) (*model.PullRequest, error)
	ListPR(context.Context, model.UserId) ([]model.PullRequest, error)
}

type TeamRepo interface {
	GetTeam(context.Context, model.TeamName) (*model.Team, error)
}

type UserRepo interface {
	GetUser(context.Context, model.UserId) (*model.User, error)
}
