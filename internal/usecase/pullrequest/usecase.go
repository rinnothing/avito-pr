package pullrequest

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/rinnothing/avito-pr/internal/model"
	"github.com/rinnothing/avito-pr/pkg/transaction"
)

const ReviewersNum = 2

type Usecase interface {
	CreatePR(context.Context, *model.PullRequest) (*model.PullRequest, error)
	MergePR(context.Context, model.PRId) (*model.PullRequest, error)
	Reassign(context.Context, model.PRId, model.UserId) (*model.PullRequest, model.UserId, error)
	ListPR(context.Context, model.UserId) (model.UserId, []model.PullRequest, error)
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

var _ Usecase = &impl{}

type impl struct {
	prR   PRRepo
	teamR TeamRepo
	userR UserRepo
}

func New(pr PRRepo, team TeamRepo, user UserRepo) *impl {
	return &impl{prR: pr, teamR: team, userR: user}
}

func (u *impl) CreatePR(ctx context.Context, pr *model.PullRequest) (*model.PullRequest, error) {
	err := transaction.DoAtomically(func(ctx context.Context) error {
		user, err := u.userR.GetUser(ctx, pr.AuthorId)
		if err != nil {
			return err
		}

		team, err := u.teamR.GetTeam(ctx, user.TeamName)
		if err != nil {
			return err
		}

		reviewers := make([]model.UserId, 0, 2)
		for _, id := range team.Members {
			if id == pr.AuthorId {
				continue
			}

			user, err = u.userR.GetUser(ctx, id)
			if err != nil {
				return err
			}

			if !user.IsActive {
				continue
			}

			reviewers = append(reviewers, id)
			if len(reviewers) == ReviewersNum {
				break
			}
		}
		pr.Reviewers = reviewers
		pr.Status = model.PROpen

		err = u.prR.CreatePR(ctx, pr)
		if errors.Is(err, model.ErrAlreadyExists) {
			err = u.prR.UpdatePR(ctx, pr)
		}
		return err
	})

	return pr, err
}

func (u *impl) MergePR(ctx context.Context, id model.PRId) (*model.PullRequest, error) {
	var pr *model.PullRequest
	err := transaction.DoAtomically(func(ctx context.Context) error {
		var err error
		pr, err = u.prR.GetPR(ctx, id)
		if err != nil {
			return err
		}

		pr.Status = model.PRMerged
		return u.prR.UpdatePR(ctx, pr)
	})

	return pr, err
}

func (u *impl) Reassign(ctx context.Context, id model.PRId, userId model.UserId) (*model.PullRequest, model.UserId, error) {
	var newUserId model.UserId
	var pr *model.PullRequest
	err := transaction.DoAtomically(func(ctx context.Context) error {
		var err error
		pr, err = u.prR.GetPR(ctx, id)
		if err != nil {
			return err
		}

		if !slices.ContainsFunc(pr.Reviewers, func(id model.UserId) bool { return id == userId }) {
			return fmt.Errorf("user with id = %s is not a reviewer of pr with id = %s: %w", userId, id, model.ErrWrongInvariant)
		}

		wrongUsers := map[model.UserId]struct{}{
			pr.AuthorId: struct{}{},
		}
		for _, revId := range pr.Reviewers {
			wrongUsers[revId] = struct{}{}
		}

		oldRev, err := u.userR.GetUser(ctx, userId)
		if err != nil {
			return err
		}

		team, err := u.teamR.GetTeam(ctx, oldRev.TeamName)
		if err != nil {
			return err
		}

		var user *model.User
		for _, newId := range team.Members {
			if _, ok := wrongUsers[newId]; ok {
				continue
			}

			user, err = u.userR.GetUser(ctx, newId)
			if err != nil {
				return err
			}

			if !user.IsActive {
				continue
			}

			newUserId = user.Id

			return nil
		}
		return fmt.Errorf("can't reassign user: no other users found in team: %w", model.ErrWrongInvariant)
	})

	for i := range pr.Reviewers {
		if pr.Reviewers[i] == userId {
			pr.Reviewers[i] = newUserId
		}
	}

	return pr, newUserId, err
}

func (u *impl) ListPR(ctx context.Context, userId model.UserId) (model.UserId, []model.PullRequest, error) {
	prs, err := u.prR.ListPR(ctx, userId)

	return userId, prs, err
}
