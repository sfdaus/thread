package pgsql

import (
	"context"
	"database/sql"
	"prakarsa-app/domain"
)

type pgsqlThreadRepository struct {
	db *sql.DB
}

// NewPgsqlThreadRepository will create new an todoRepository object representation of ThreadRepository interface
func NewPgsqlThreadRepository(db *sql.DB) *pgsqlThreadRepository {
	return &pgsqlThreadRepository{
		db: db,
	}
}

func (r *pgsqlThreadRepository) Create(ctx context.Context, thread *domain.Thread) (err error) {
	query := "INSERT INTO threads (name, created_at, updated_at) VALUES ($1, $2, $3)"
	_, err = r.db.ExecContext(ctx, query, thread.Name, thread.CreatedAt, thread.UpdatedAt)
	return
}
