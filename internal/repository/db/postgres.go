package db

import (
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	uniqueViolation     = "23505"
	foreignKeyViolation = "23503"
)

type postgresRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRepository(db *pgxpool.Pool) *postgresRepository {
	return &postgresRepository{
		db: db,
	}
}
