package db

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"github.com/rinnothing/avito-pr/internal/model"
	"github.com/rinnothing/avito-pr/internal/usecase/pullrequest"
	"github.com/rinnothing/avito-pr/internal/usecase/team"
	"github.com/rinnothing/avito-pr/pkg/transaction"
)

var _ team.TeamRepo = &postgresRepository{}
var _ pullrequest.TeamRepo = &postgresRepository{}

func (p *postgresRepository) CreateTeam(ctx context.Context, team *model.Team) (txErr error) {
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

	const queryTeam = `
INSERT INTO teams (name)
VALUES ($1)
ON CONFLICT (name) DO NOTHING
RETURNING name
`

	var createdName model.TeamName
	err = tx.QueryRow(ctx, queryTeam, team.Name).Scan(&createdName)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.ErrAlreadyExists
		}
		return err
	}

	return nil
}

func (p *postgresRepository) GetTeam(ctx context.Context, name model.TeamName) (retTeam *model.Team, txErr error) {
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

	const queryTeam = `
SELECT name FROM teams WHERE name = $1
`

	var scannedName model.TeamName
	err = tx.QueryRow(ctx, queryTeam, name).Scan(&scannedName)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, err
	}

	const queryUser = `
SELECT id FROM users WHERE team_name = $1
`

	rows, err := tx.Query(ctx, queryUser, name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ids := make([]model.UserId, 0)
	for rows.Next() {
		var id model.UserId
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}

		ids = append(ids, id)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	team := model.Team{
		Name:    name,
		Members: ids,
	}
	return &team, nil
}
