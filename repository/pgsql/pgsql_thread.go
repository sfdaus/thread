package pgsql

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"prakarsa-app/entity"
	"prakarsa-app/transport/request"
	"prakarsa-app/transport/response"
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

func (r *pgsqlThreadRepository) Create(ctx context.Context, thread *entity.Thread, attachments []*entity.Attachment, tags []*entity.ThreadTag,
	institutions []*entity.ThreadInstitution, partnerTypes []*entity.ThreadPartnerType) (err error) {
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
				status, created_by, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11,$12)`
	if _, err = r.db.ExecContext(ctx, query, thread.ID, thread.UserID, thread.Title, pq.Array(thread.Type), thread.Description, thread.UpvoteNumber, thread.ReportNumber,
		thread.FollowedNumber, thread.Status, thread.CreatedBy, thread.CreatedAt, thread.CreatedAt); err != nil {
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
func (r *pgsqlThreadRepository) Update(ctx context.Context, thread *entity.Thread, attachments []*entity.Attachment, removedAttachments []string) (err error) {
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

func (r *pgsqlThreadRepository) Delete(ctx context.Context, thread *entity.Thread) (rowsAffected int64, err error) {
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

func (r *pgsqlThreadRepository) GetList(ctx context.Context, request *request.GetListThreadReq) (threads []response.GetListThreadTempRes, meta response.MetaRes, err error) {
	// 1. Build WHERE clauses
	wheres := []string{}
	args := []interface{}{}
	idx := 1

	if request.Title != "" {
		wheres = append(wheres, fmt.Sprintf("t.title ILIKE $%d", idx))
		args = append(args, "%"+request.Title+"%")
		idx++
	}

	if request.Status != "" {
		wheres = append(wheres, fmt.Sprintf("t.status = $%d", idx))
		args = append(args, request.Status)
		idx++
	}

	if request.IsActive != nil {
		wheres = append(wheres, fmt.Sprintf("t.is_active = $%d", idx))
		args = append(args, request.IsActive)
		idx++
	}

	whereSQL := ""
	if len(wheres) > 0 {
		whereSQL = "WHERE " + strings.Join(wheres, " AND ")
	}

	// --- 2. Hitung totalCount dulu (tanpa LIMIT/OFFSET) ---
	countQuery := fmt.Sprintf(
		`SELECT COUNT(*) 
				FROM threads as t %s`,
		whereSQL,
	)
	if err = r.db.QueryRowContext(ctx, countQuery, args...).Scan(&meta.TotalData); err != nil {
		return nil, meta, err
	}

	// 2. Calculate LIMIT & OFFSET
	perPage := request.PerPage
	if perPage <= 0 {
		perPage = 10
	}
	page := request.Page
	if page <= 0 {
		page = 1
	}

	// total pages = ceil(total / perPage)
	meta.Page = page
	meta.PerPage = perPage
	meta.TotalPages = (meta.TotalData + perPage - 1) / perPage

	offset := (page - 1) * perPage

	// add LIMIT & OFFSET to args
	args = append(args, perPage, offset)
	limitPos, offsetPos := idx, idx+1

	// 3. Final query
	query := fmt.Sprintf(`
        SELECT
            t.id, t.user_id, t.title, t.type, t.description, t.status,
		  t.upvote_number, t.report_number, t.followed_number, COALESCE(t.deadline, NULL) AS deadline,
		  t.is_active, t.created_by, t.created_at, t.updated_by, t.updated_at, t.deleted_at,
		
		  COALESCE(
			jsonb_agg(
			  DISTINCT jsonb_build_object(
				'id', ta.id, 'thread_id', ta.thread_id,
				'file_name', ta.file_name, 'file_url', ta.file_url, 'file_type', ta.file_type,
				'is_active', ta.is_active, 'created_by', ta.created_by, 'created_at', ta.created_at,
				'updated_by', ta.updated_by, 'updated_at', ta.updated_at, 'deleted_at', ta.deleted_at
			  )
			) FILTER (WHERE ta.id IS NOT NULL),
			'[]'::jsonb
		  ) AS attachments,
		
		  COALESCE(
			jsonb_agg(
			  DISTINCT jsonb_build_object(
				'id', tt.id, 'thread_id', tt.thread_id, 'tag_id', tt.tag_id,
				'is_active', tt.is_active, 'created_by', tt.created_by, 'created_at', tt.created_at,
				'updated_by', tt.updated_by, 'updated_at', tt.updated_at, 'deleted_at', tt.deleted_at
			  )
			) FILTER (WHERE tt.id IS NOT NULL),
			'[]'::jsonb
		  ) AS tags,
		
		  COALESCE(
			jsonb_agg(
			  DISTINCT jsonb_build_object(
				'id', tpt.id, 'thread_id', tpt.thread_id, 'partner_type_id', tpt.partner_type_id,
				'compensation_type', tpt.compensation_type, 'compensation_value', tpt.compensation_value,
				'compensation_currency', tpt.compensation_currency, 'compensation_period', tpt.compensation_period,
				'compensation_note', tpt.compensation_note,
				'is_active', tpt.is_active, 'created_by', tpt.created_by, 'created_at', tpt.created_at,
				'updated_by', tpt.updated_by, 'updated_at', tpt.updated_at, 'deleted_at', tpt.deleted_at
			  )
			) FILTER (WHERE tpt.id IS NOT NULL),
			'[]'::jsonb
		  ) AS partner_types,
		
		  COALESCE(
			jsonb_agg(
			  DISTINCT jsonb_build_object(
				'id', ti.id, 'thread_id', ti.thread_id, 'institution_id', ti.institution_id,
				'is_active', ti.is_active, 'created_by', ti.created_by, 'created_at', ti.created_at,
				'updated_by', ti.updated_by, 'updated_at', ti.updated_at, 'deleted_at', ti.deleted_at
			  )
			) FILTER (WHERE ti.id IS NOT NULL),
			'[]'::jsonb
		  ) AS institutions
        FROM threads as t
        LEFT JOIN thread_attachments   AS ta  ON ta.thread_id  = t.id AND ta.is_active  = true
		LEFT JOIN thread_institutions  AS ti  ON ti.thread_id  = t.id AND ti.is_active  = true
		LEFT JOIN thread_tags          AS tt  ON tt.thread_id  = t.id AND tt.is_active  = true
		LEFT JOIN thread_partner_types AS tpt ON tpt.thread_id = t.id AND tpt.is_active = true
        %s
       GROUP BY
		  t.id, t.user_id, t.title, t.type, t.description, t.status,
		  t.upvote_number, t.report_number, t.followed_number, t.deadline,
		  t.is_active, t.created_by, t.created_at, t.updated_by, t.updated_at, t.deleted_at
		ORDER BY t.created_at DESC
        LIMIT $%d OFFSET $%d`, whereSQL, limitPos, offsetPos)

	// 4. Execute
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, meta, err
	}
	defer rows.Close()

	// 5. Scan results
	type listRow struct {
		entity.Thread

		AttachmentsJSON  []byte
		TagsJSON         []byte
		PartnerTypesJSON []byte
		InstitutionsJSON []byte
	}

	for rows.Next() {
		var r listRow

		// Tambah holder untuk kolom yang nullable
		var (
			deadlineNT  sql.NullTime
			updatedByNS sql.NullString
			updatedAtNT sql.NullInt64
			deletedAtNT sql.NullInt64
		)

		if err := rows.Scan(
			&r.ID, &r.UserID, &r.Title, pq.Array(&r.Type), &r.Description, &r.Status,
			&r.UpvoteNumber, &r.ReportNumber, &r.FollowedNumber, &deadlineNT, // <— pakai NullTime
			&r.IsActive, &r.CreatedBy, &r.CreatedAt, &updatedByNS, &updatedAtNT, &deletedAtNT, // <— NullString/NullTime
			&r.AttachmentsJSON, &r.TagsJSON, &r.PartnerTypesJSON, &r.InstitutionsJSON,
		); err != nil {
			return nil, meta, err
		}

		// Map balik ke tipe di entity.Thread
		if deadlineNT.Valid {
			r.Deadline = &deadlineNT.Time
		} else {
			r.Deadline = nil
		}
		
		if updatedByNS.Valid {
			r.UpdatedBy = updatedByNS.String
		} else {
			r.UpdatedBy = "" // atau "system"
		}

		if updatedAtNT.Valid {
			r.UpdatedAt = updatedAtNT.Int64
		} else {
			r.UpdatedAt = 0
		}
		if deletedAtNT.Valid {
			r.DeletedAt = deletedAtNT.Int64
		} else {
			r.DeletedAt = 0
		}

		var res response.GetListThreadTempRes
		res.Thread = r.Thread

		if err := json.Unmarshal(r.AttachmentsJSON, &res.Attachments); err != nil {
			return nil, meta, err
		}
		if err := json.Unmarshal(r.TagsJSON, &res.Tags); err != nil {
			return nil, meta, err
		}
		if err := json.Unmarshal(r.PartnerTypesJSON, &res.PartnerTypes); err != nil {
			return nil, meta, err
		}
		if err := json.Unmarshal(r.InstitutionsJSON, &res.Institutions); err != nil {
			return nil, meta, err
		}

		threads = append(threads, res)
	}

	if errRow := rows.Err(); errRow != nil {
		return nil, meta, errRow
	}

	return
}
