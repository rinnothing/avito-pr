package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/rinnothing/avito-pr/internal/model"
	"github.com/rinnothing/avito-pr/internal/usecase/pullrequest"
	"github.com/rinnothing/avito-pr/pkg/transaction"
)

var _ pullrequest.PRRepo = &postgresRepository{}

func (p *postgresRepository) CreatePR(ctx context.Context, pr *model.PullRequest) (txErr error) {
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

	const queryPR = `
INSERT INTO pull_requests (id, name, author_id, status, created_at, merged_at)
VALUES ($1, $2, $3, 'open', NOW(), NULL);
`

	_, err = tx.Exec(ctx, queryPR, pr.Id, pr.Name, pr.AuthorId)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			switch pgErr.Code {
			case uniqueViolation:
				return model.ErrAlreadyExists
			case foreignKeyViolation:
				return model.ErrNotFound
			}
		}
		return err
	}

	const queryUsers = `
INSERT INTO pr_reviewers (pr_id, user_id)
SELECT $1, unnest($2::text[])
ON CONFLICT DO NOTHING
`

	if len(pr.Reviewers) > 0 {
		_, err := tx.Exec(ctx, queryUsers, pr.Id, pr.Reviewers)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *postgresRepository) UpdatePR(ctx context.Context, pr *model.PullRequest) (txErr error) {
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

	const queryPR = `
UPDATE pull_requests
SET name = $2,
    status = $3,
    merged_at = $4
WHERE id = $1
RETURNING id
`

	var newId model.PRId
	mergedAt := (*time.Time)(nil)
	if pr.MergedAt != nil {
		mergedAt = pr.MergedAt
	}

	var internalStatus string
	switch pr.Status {
	case model.PROpen:
		internalStatus = "open"
	case model.PRMerged:
		internalStatus = "merged"
	default:
		return fmt.Errorf("no such model status: %d", pr.Status)
	}
	err = tx.QueryRow(ctx, queryPR, pr.Id, pr.Name, internalStatus, mergedAt).Scan(&newId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.ErrNotFound
		}
		return err
	}

	const queryDelete = `
DELETE FROM pr_reviewers
WHERE pr_id = $1
AND user_id <> ALL($2)
`
	_, err = tx.Exec(ctx, queryDelete, pr.Id, pr.Reviewers)
	if err != nil {
		return err
	}

	const queryInsert = `
INSERT INTO pr_reviewers (pr_id, user_id)
SELECT $1, unnest($2::text[])
ON CONFLICT (pr_id, user_id) DO NOTHING
`

	if len(pr.Reviewers) > 0 {
		_, err = tx.Exec(ctx, queryInsert, pr.Id, pr.Reviewers)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *postgresRepository) GetPR(ctx context.Context, id model.PRId) (retPr *model.PullRequest, txErr error) {
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

	const queryPR = `
SELECT id, name, author_id, status, created_at, merged_at
FROM pull_requests
WHERE id = $1
`

	var pr model.PullRequest
	var status string
	err = tx.QueryRow(ctx, queryPR, id).Scan(&pr.Id, &pr.Name, &pr.AuthorId, &status, &pr.CreatedAt, &pr.MergedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, err
	}

	switch status {
	case "open":
		pr.Status = model.PROpen
	case "merged":
		pr.Status = model.PRMerged
	default:
		return nil, fmt.Errorf("unknown status: %s", status)
	}

	const queryUser = `
SELECT user_id FROM pr_reviewers WHERE pr_id = $1
`
	rows, err := tx.Query(ctx, queryUser, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	pr.Reviewers = make([]model.UserId, 0)
	for rows.Next() {
		var userId model.UserId
		if err = rows.Scan(&userId); err != nil {
			return nil, err
		}
		pr.Reviewers = append(pr.Reviewers, userId)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return &pr, nil
}

func (p *postgresRepository) ListPR(ctx context.Context, id model.UserId) (retPRs []model.PullRequest, txErr error) {
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
SELECT pr_id FROM pr_reviewers WHERE user_id = $1`

	rows, err := tx.Query(ctx, queryUser, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prIds []model.PRId
	for rows.Next() {
		var id model.PRId
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		prIds = append(prIds, id)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	resPrs := make([]model.PullRequest, 0)
	for _, prId := range prIds {
		pr, err := p.GetPR(ctx, prId)
		if err != nil {
			return nil, err
		}

		resPrs = append(resPrs, *pr)
	}

	return resPrs, nil
}
