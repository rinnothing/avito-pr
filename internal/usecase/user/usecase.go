package user

import (
	"context"

	"github.com/rinnothing/avito-pr/internal/model"
	"github.com/rinnothing/avito-pr/pkg/transaction"
)

type Usecase interface {
	SetIsActive(context.Context, model.UserId, bool) (*model.User, error)
}

type UserRepo interface {
	GetUser(context.Context, model.UserId) (*model.User, error)
	UpdateUser(context.Context, *model.User) error
}

var _ Usecase = &impl{}

type impl struct {
	repo UserRepo
	tr   transaction.Transactor
}

func New(repo UserRepo, tr transaction.Transactor) *impl {
	return &impl{repo: repo, tr: tr}
}

func (u *impl) SetIsActive(ctx context.Context, id model.UserId, isActive bool) (*model.User, error) {
	var user *model.User
	err := u.tr.DoAtomically(ctx, func(ctx context.Context) error {
		var err error
		user, err = u.repo.GetUser(ctx, id)
		if err != nil {
			return err
		}

		if user.IsActive == isActive {
			return nil
		}
		user.IsActive = isActive
		return u.repo.UpdateUser(ctx, user)
	})

	return user, err
}
