package pgsql

import (
	"context"
	"database/sql"
	"prakarsa-app/domain"

	"github.com/lib/pq"
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

func (r *pgsqlThreadRepository) Create(ctx context.Context, thread *domain.Thread, attachments []*domain.Attachment) (err error) {
	// Mulai transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	// Pastikan rollback kalau ada error
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	query := `INSERT INTO threads (id, user_id, title, type, description, upvote_number, report_number, followed_number, 
				status, created_by, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`
	if _, err = r.db.ExecContext(ctx, query, thread.ID, thread.UserID, thread.Title, pq.Array(thread.Type), thread.Description, thread.UpvoteNumber, thread.ReportNumber,
		thread.FollowedNumber, thread.Status, thread.CreatedBy, thread.CreatedAt); err != nil {
		return err
	}

	if len(attachments) > 0 {
		for _, attachment := range attachments {
			query := `INSERT INTO thread_attachments (id, thread_id, file_url, file_type, file_name, is_active, created_by, created_at) 
						VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
			if _, err = r.db.ExecContext(ctx, query, attachment.ID, attachment.ThreadID, attachment.FileUrl, attachment.FileType, attachment.FileName,
				attachment.IsActive, attachment.CreatedBy, attachment.CreatedAt); err != nil {
				return err
			}
		}
	}

	// Commit jika semua sukses
	if err = tx.Commit(); err != nil {
		return err
	}

	return
}
