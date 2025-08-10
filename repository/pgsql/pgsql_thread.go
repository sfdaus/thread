package pgsql

import (
	"context"
	"database/sql"
	"fmt"
	"prakarsa-app/domain"
	"prakarsa-app/transport/request"
	"prakarsa-app/utils"
	"strings"

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

func (r *pgsqlThreadRepository) Create(ctx context.Context, thread *domain.Thread, attachments []*domain.Attachment, tags []*domain.ThreadTag,
	institutions []*domain.ThreadInstitution, partnerTypes []*domain.ThreadPartnerType) (err error) {
	// Mulai transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	// Pastikan rollback kalau ada error
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
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
		query = `INSERT INTO thread_attachments (id, thread_id, file_url, file_type, file_name, is_active, created_by, created_at) 
						VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
		for _, attachment := range attachments {
			if _, err = r.db.ExecContext(ctx, query, attachment.ID, attachment.ThreadID, attachment.FileUrl, attachment.FileType, attachment.FileName,
				attachment.IsActive, attachment.CreatedBy, attachment.CreatedAt); err != nil {
				return err
			}
		}
	}

	if len(tags) > 0 {
		const query = `INSERT INTO thread_tags (id, thread_id, tag_id, is_active, created_at, created_by)
			VALUES ($1,$2,$3,$4,$5,$6)`

		for _, tag := range tags {
			if _, err = r.db.ExecContext(ctx, query, tag.ID, tag.ThreadID, tag.TagID,
				tag.IsActive, tag.CreatedAt, tag.CreatedBy); err != nil {
				return err
			}
		}
	}

	if len(institutions) > 0 {
		const query = `INSERT INTO thread_institutions (id, thread_id, institution_id, is_active, created_by, 
			created_at) VALUES ($1,$2,$3,$4,$5,$6)`

		for _, institution := range institutions {
			if _, err = tx.ExecContext(ctx, query,
				institution.ID, institution.ThreadID, institution.InstitutionID,
				institution.IsActive, institution.CreatedBy, institution.CreatedAt,
			); err != nil {
				return err
			}
		}
	}

	if len(partnerTypes) > 0 {
		const query = `INSERT INTO thread_partner_types (
							id, thread_id, partner_type_id,
							compensation_type, compensation_value, compensation_currency, compensation_period, compensation_note,
							is_active, created_by, created_at
						) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`
		for _, p := range partnerTypes {
			if _, err = tx.ExecContext(ctx, query,
				p.ID, p.ThreadID, p.PartnerTypeID,
				p.CompensationType, p.CompensationValue, p.CompensationCurrency, p.CompensationPeriod, p.CompensationNote,
				p.IsActive, p.CreatedBy, p.CreatedAt,
			); err != nil {
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
func (r *pgsqlThreadRepository) Update(ctx context.Context, thread *domain.Thread, attachments []*domain.Attachment, removedAttachments []string) (err error) {
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

	// Build dynamic SET clauses from Thread struct
	sets := []string{fmt.Sprintf("updated_at = $%d", 1), fmt.Sprintf("updated_by = $%d", 2)}
	args := []interface{}{thread.UpdatedAt, thread.UpdatedBy}
	idx := 3

	if thread.Title != "" {
		sets = append(sets, fmt.Sprintf("title = $%d", idx))
		args = append(args, thread.Title)
		idx++
	}
	if len(thread.Type) > 0 {
		sets = append(sets, fmt.Sprintf("type = $%d", idx))
		args = append(args, pq.Array(thread.Type))
		idx++
	}
	if thread.Description != "" {
		sets = append(sets, fmt.Sprintf("description = $%d", idx))
		args = append(args, thread.Description)
		idx++
	}
	if thread.Status != "" {
		sets = append(sets, fmt.Sprintf("status = $%d", idx))
		args = append(args, thread.Status)
		idx++
	}

	if !thread.Deadline.IsZero() {
		sets = append(sets, fmt.Sprintf("deadline = $%d", idx))
		args = append(args, thread.Deadline)
		idx++
	}

	// kalau ada sesuatu untuk di‐update, commit ke SQL
	if len(sets) > 0 {
		// tambahkan WHERE id = $idx
		args = append(args, thread.ID)
		query := fmt.Sprintf(
			"UPDATE threads SET %s WHERE id = $%d",
			strings.Join(sets, ", "),
			idx,
		)
		if _, err = tx.ExecContext(ctx, query, args...); err != nil {
			return
		}
	}

	// huru hara attachment
	var existingCount int
	err = tx.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM thread_attachments WHERE thread_id = $1`,
		thread.ID,
	).Scan(&existingCount)
	if err != nil {
		return
	}

	if newAttachCount := existingCount - len(removedAttachments) + len(attachments); newAttachCount > utils.MaxTotalAttachments {
		err = utils.NewBadRequestError("Maximum number of attachments exceeded.")
		return
	}

	if len(removedAttachments) > 0 {
		_, err = tx.ExecContext(ctx,
			`DELETE FROM thread_attachments
               WHERE thread_id = $1
                 AND id = ANY($2)`,
			thread.ID, pq.Array(removedAttachments),
		)
		if err != nil {
			return
		}
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

func (r *pgsqlThreadRepository) Delete(ctx context.Context, thread *domain.Thread) (rowsAffected int64, err error) {
	query := "DELETE FROM threads WHERE id = $1"
	res, err := r.db.ExecContext(ctx, query, thread.ID)
	if err != nil {
		return
	}

	rowsAffected, err = res.RowsAffected()
	if err != nil {
		return
	}

	return
}

func (r *pgsqlThreadRepository) GetList(ctx context.Context, request *request.GetListThreadReq) {
	return
}
