package user

import (
	"context"

	"github.com/rinnothing/avito-pr/internal/model"
)

type Usecase interface {
	SetIsActive(context.Context, model.UserId, bool) (*model.User, error)
}

type UserRepo interface {
	GetUser(context.Context, model.UserId) (*model.User, error)
	UpdateUser(context.Context, *model.User) error
}
