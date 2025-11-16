package db

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"github.com/rinnothing/avito-pr/internal/model"
	"github.com/rinnothing/avito-pr/internal/usecase/pullrequest"
	"github.com/rinnothing/avito-pr/internal/usecase/team"
	"github.com/rinnothing/avito-pr/internal/usecase/user"
	"github.com/rinnothing/avito-pr/pkg/transaction"
)

var _ user.UserRepo = &postgresRepository{}
var _ team.UserRepo = &postgresRepository{}
var _ pullrequest.UserRepo = &postgresRepository{}

func (p *postgresRepository) CreateUser(ctx context.Context, user *model.User) (txErr error) {
	// this thing is actually a copypaste from my older projects
	var (
		tx  pgx.Tx
		err error
	)
	if tx, err = transaction.ExtractTx(ctx); err != nil {
		tx, err = p.db.Begin(ctx)
		if err != nil {
			return err
		}

		defer func() {
			if txErr != nil {
				_ = tx.Rollback(ctx)
				return
			}

			_ = tx.Commit(ctx)
		}()
	}

	const queryUser = `
INSERT INTO users (id, username, team_name, is_active)
VALUES ($1, $2, $3, $4)
ON CONFLICT (id) DO NOTHING
RETURNING id
`

	var insertedID string
	err = tx.QueryRow(ctx, queryUser, user.Id, user.Username, user.TeamName, user.IsActive).Scan(&insertedID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.ErrAlreadyExists
		}
		return err
	}
	return nil
}

func (p *postgresRepository) UpdateUser(ctx context.Context, user *model.User) (txErr error) {
	var (
		tx  pgx.Tx
		err error
	)
	if tx, err = transaction.ExtractTx(ctx); err != nil {
		tx, err = p.db.Begin(ctx)
		if err != nil {
			return err
		}

		defer func() {
			if txErr != nil {
				_ = tx.Rollback(ctx)
				return
			}

			_ = tx.Commit(ctx)
		}()
	}

	const queryRow = `
UPDATE users
SET username = $2,
    team_name = $3,
    is_active = $4
WHERE id = $1
RETURNING id
`

	var insertedID string
	err = tx.QueryRow(ctx, queryRow, user.Id, user.Username, user.TeamName, user.IsActive).Scan(insertedID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.ErrNotFound
		}
		return err
	}

	return nil
}

func (p *postgresRepository) GetUser(ctx context.Context, id model.UserId) (res *model.User, txErr error) {
	var (
		tx  pgx.Tx
		err error
	)
	if tx, err = transaction.ExtractTx(ctx); err != nil {
		tx, err = p.db.Begin(ctx)
		if err != nil {
			return nil, err
		}

		defer func() {
			if txErr != nil {
				_ = tx.Rollback(ctx)
				return
			}

			_ = tx.Commit(ctx)
		}()
	}

	const queryUser = `
SELECT id, username, team_name, is_active
FROM users
WHERE id = $1
`

	var user model.User
	err = tx.QueryRow(ctx, queryUser, id).Scan(&user.Id, &user.Username, &user.TeamName, &user.IsActive)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, err
	}

	return &user, nil
}
