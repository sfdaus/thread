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
							amount_needed, is_active, created_by, created_at
						) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11, $12)`
		for _, p := range partnerTypes {
			if _, err = tx.ExecContext(ctx, query,
				p.ID, p.ThreadID, p.PartnerTypeID,
				p.CompensationType, p.CompensationValue, p.CompensationCurrency, p.CompensationPeriod, p.CompensationNote,
				p.AmountNeeded, p.IsActive, p.CreatedBy, p.CreatedAt,
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
func (r *pgsqlThreadRepository) Update(ctx context.Context, thread *entity.Thread, attachments []*entity.Attachment, removedAttachments []string,
	addedTags []*entity.ThreadTag, removedTags []string, addedInstitutions []*entity.ThreadInstitution, removedInstitutions []string,
	partnerTypes []*entity.UpdateThreadPartnerType, excludeRemovePartnerTypes []string) (err error) {

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

	if thread.Deadline != nil {
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
			query := `INSERT INTO thread_attachments (id, thread_id, file_url, file_type, file_name, is_active, created_by, created_at, updated_at, updated_by)
						VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`
			if _, err = tx.ExecContext(ctx, query, attachment.ID, attachment.ThreadID, attachment.FileUrl, attachment.FileType, attachment.FileName,
				attachment.IsActive, attachment.CreatedBy, attachment.CreatedAt, attachment.UpdatedAt, attachment.UpdatedBy); err != nil {
				return err
			}
		}
	}

	/*
		Tags
	*/
	if len(removedTags) > 0 {
		_, err = tx.ExecContext(ctx,
			`DELETE FROM thread_tags
               WHERE thread_id = $1
                 AND id = ANY($2)`,
			thread.ID, pq.Array(removedTags),
		)
		if err != nil {
			return
		}
	}

	if len(addedTags) > 0 {
		for _, tag := range addedTags {
			query := `INSERT INTO thread_tags (id, thread_id, tag_id, is_active, created_by, created_at, updated_at, updated_by)
						VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
			if _, err = tx.ExecContext(ctx, query, tag.ID, tag.ThreadID, tag.TagID, tag.IsActive, tag.CreatedBy,
				tag.CreatedAt, tag.UpdatedAt, tag.UpdatedBy); err != nil {
				return err
			}
		}
	}

	/*
		Institutions
	*/
	if len(removedInstitutions) > 0 {
		_, err = tx.ExecContext(ctx,
			`DELETE FROM thread_institutions
               WHERE thread_id = $1
                 AND id = ANY($2)`,
			thread.ID, pq.Array(removedInstitutions),
		)
		if err != nil {
			return
		}
	}

	if len(addedInstitutions) > 0 {
		for _, institution := range addedInstitutions {
			query := `INSERT INTO thread_institutions (id, thread_id, institution_id, is_active, created_by, created_at, updated_at, updated_by)
						VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
			if _, err = tx.ExecContext(ctx, query, institution.ID, institution.ThreadID, institution.InstitutionID, institution.IsActive, institution.CreatedBy,
				institution.CreatedAt, institution.UpdatedAt, institution.UpdatedBy); err != nil {
				return err
			}
		}
	}

	/*
		Partner Types
	*/
	if len(excludeRemovePartnerTypes) > 0 {
		vals := []any{thread.ID}
		ph := make([]string, len(excludeRemovePartnerTypes))

		for i, v := range excludeRemovePartnerTypes {
			ph[i] = fmt.Sprintf("$%d", i+2)
			vals = append(vals, v)
		}

		query := fmt.Sprintf(`DELETE FROM thread_partner_types t
	 		WHERE t.thread_id=$1 AND t.id NOT IN (%s)`, strings.Join(ph, ","))
		if _, err := tx.ExecContext(ctx, query,
			vals...,
		); err != nil {
			return err
		}
	}

	if len(partnerTypes) > 0 {
		query := `INSERT INTO thread_partner_types(id, thread_id, partner_type_id,
					 compensation_type, compensation_value, compensation_currency,
					 compensation_period, compensation_note, amount_needed,
					 created_at, created_by, updated_at, updated_by)
					VALUES
					($1,$2,COALESCE($3,(SELECT partner_type_id FROM thread_partner_types WHERE id = $14)),
					 $4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
					ON CONFLICT (id) 
					DO UPDATE 
					SET
				  partner_type_id=COALESCE(EXCLUDED.partner_type_id,thread_partner_types.partner_type_id),
				  compensation_type=COALESCE(EXCLUDED.compensation_type,thread_partner_types.compensation_type),
				  compensation_value=COALESCE(EXCLUDED.compensation_value,thread_partner_types.compensation_value),
				  compensation_currency=COALESCE(EXCLUDED.compensation_currency,thread_partner_types.compensation_currency),
				  compensation_period=COALESCE(EXCLUDED.compensation_period,thread_partner_types.compensation_period),
				  compensation_note=COALESCE(EXCLUDED.compensation_note,thread_partner_types.compensation_note),
				  amount_needed=COALESCE(EXCLUDED.amount_needed,thread_partner_types.amount_needed),
				  updated_at=EXCLUDED.updated_at,
				  updated_by=EXCLUDED.updated_by`

		for _, partnerType := range partnerTypes {
			if _, err = tx.ExecContext(ctx, query,
				partnerType.ID, partnerType.ThreadID, partnerType.PartnerTypeID, partnerType.CompensationType,
				partnerType.CompensationValue, partnerType.CompensationCurrency, partnerType.CompensationPeriod,
				partnerType.CompensationNote, partnerType.AmountNeeded, partnerType.CreatedAt, partnerType.CreatedBy,
				partnerType.UpdatedAt, partnerType.UpdatedBy, partnerType.ID,
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

	// 3. total pages = ceil(total / perPage)
	meta.Page = page
	meta.PerPage = perPage
	meta.TotalPages = (meta.TotalData + perPage - 1) / perPage

	offset := (page - 1) * perPage

	// add LIMIT & OFFSET to args
	args = append(args, perPage, offset)
	limitPos, offsetPos := idx, idx+1

	// 4. Data query (Profile + Institution in Profile)
	query := fmt.Sprintf(`
		SELECT
		  -- thread
		  t.id, t.user_id, t.title, t.type, t.description, t.status,
		  t.upvote_number, t.report_number, t.followed_number, t.deadline,
		  t.is_active, t.created_by, t.created_at, t.updated_by, t.updated_at, t.deleted_at,

		  -- profile (1-1)
		  COALESCE(p.name,'')        AS prof_name,
		  COALESCE(p.name_alias,'')  AS prof_name_alias,
		  COALESCE(p.avatar,'') AS prof_avatar,

		  -- institution inside profile
		  COALESCE(i.name,'')  AS prof_inst_name,
		  COALESCE(i.alias,'') AS prof_inst_alias,   -- hapus baris ini kalau kolom alias belum ada di DB
		  COALESCE(i.type,'')  AS prof_inst_type,

		  -- aggregates
		  COALESCE(ja.attachments,'[]'::jsonb)   AS attachments,
		  COALESCE(jtg.tags,'[]'::jsonb)         AS tags,
		  COALESCE(jpt.partner_types,'[]'::jsonb) AS partner_types,
		  COALESCE(ji.institutions,'[]'::jsonb)  AS institutions
		FROM threads t
		LEFT JOIN profiles p     ON p.user_id = t.user_id
		LEFT JOIN institutions i ON i.id = p.institution_id

		-- attachments
		LEFT JOIN LATERAL (
		  SELECT jsonb_agg(
				   jsonb_build_object(
					 'id', ta.id,
					 'file_name', ta.file_name, 'file_url', ta.file_url, 'file_type', ta.file_type,
					 'is_active', ta.is_active, 'created_at', ta.created_at,
					 'updated_at', ta.updated_at
				   ) ORDER BY ta.created_at DESC
				 ) AS attachments
		  FROM thread_attachments ta
		  WHERE ta.thread_id = t.id AND ta.is_active = true
		) ja ON true

		-- tags
		LEFT JOIN LATERAL (
		  SELECT jsonb_agg(
				   jsonb_build_object(
					 'id', tg.id,
					 'name', tg.name,
					 'description', tg.description,
					 'is_active', tt.is_active, 'created_at', tt.created_at,
					 'updated_at', tt.updated_at
				   ) ORDER BY tg.name
				 ) AS tags
		  FROM thread_tags tt
		  JOIN tags tg ON tg.id = tt.tag_id
		  WHERE tt.thread_id = t.id AND tt.is_active = true
		) jtg ON true

		-- partner types
		LEFT JOIN LATERAL (
		  SELECT jsonb_agg(
				   jsonb_build_object(
					 'id', tpt.id,
					 'name', pt.name,
					 'compensation_type', ct.name,
					 'compensation_value', tpt.compensation_value,
					 'compensation_currency', tpt.compensation_currency,
					 'compensation_period', tpt.compensation_period,
					 'compensation_note', tpt.compensation_note,
					 'amount_needed', tpt.amount_needed,
					 'amount_fulfilled', tpt.amount_fulfilled,
					 'is_active', tpt.is_active,
					 'created_at', tpt.created_at,
					 'updated_at', tpt.updated_at
				   ) ORDER BY pt.name
				 ) AS partner_types
		  FROM thread_partner_types tpt
		  JOIN partner_types pt ON pt.id = tpt.partner_type_id
		  LEFT JOIN compensation_types ct ON ct.id = tpt.compensation_type
		  WHERE tpt.thread_id = t.id AND tpt.is_active = true
		) jpt ON true

		-- institutions on thread
		LEFT JOIN LATERAL (
		  SELECT jsonb_agg(
				   jsonb_build_object(
					 'id', i2.id,
					 'name', i2.name,
					 'alias', i2.alias,
					 'type', i2.type,
					 'is_active', ti.is_active, 'created_at', ti.created_at,
					 'updated_at', ti.updated_at
				   ) ORDER BY i2.name
				 ) AS institutions
		  FROM thread_institutions ti
		  JOIN institutions i2 ON i2.id = ti.institution_id
		  WHERE ti.thread_id = t.id AND ti.is_active = true
		) ji ON true

		%s
		ORDER BY t.created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereSQL, limitPos, offsetPos)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, meta, err
	}
	defer rows.Close()

	type listRow struct {
		entity.Thread
		ProfName      string
		ProfNameAlias string
		ProfAvatar    string
		ProfInstName  string
		ProfInstAlias string
		ProfInstType  string

		AttachmentsJSON  []byte
		TagsJSON         []byte
		PartnerTypesJSON []byte
		InstitutionsJSON []byte
	}

	for rows.Next() {
		var rrow listRow

		var (
			deadlineNT  sql.NullTime
			updatedByNS sql.NullString
			updatedAtNT sql.NullInt64
			deletedAtNT sql.NullInt64
		)

		if err := rows.Scan(
			&rrow.ID, &rrow.UserID, &rrow.Title, pq.Array(&rrow.Type), &rrow.Description, &rrow.Status,
			&rrow.UpvoteNumber, &rrow.ReportNumber, &rrow.FollowedNumber, &deadlineNT,
			&rrow.IsActive, &rrow.CreatedBy, &rrow.CreatedAt, &updatedByNS, &rrow.UpdatedAt, &deletedAtNT,

			&rrow.ProfName, &rrow.ProfNameAlias, &rrow.ProfAvatar,
			&rrow.ProfInstName, &rrow.ProfInstAlias, &rrow.ProfInstType,

			&rrow.AttachmentsJSON, &rrow.TagsJSON, &rrow.PartnerTypesJSON, &rrow.InstitutionsJSON,
		); err != nil {
			return nil, meta, err
		}

		// map nullable thread fields
		if deadlineNT.Valid {
			rrow.Deadline = &deadlineNT.Time
		} else {
			rrow.Deadline = nil
		}
		if updatedByNS.Valid {
			rrow.UpdatedBy = updatedByNS.String
		} else {
			rrow.UpdatedBy = ""
		}
		if !updatedAtNT.Valid {
			rrow.UpdatedAt = 0
		}
		if deletedAtNT.Valid {
			rrow.DeletedAt = deletedAtNT.Int64
		} else {
			rrow.DeletedAt = 0
		}

		var out response.GetListThreadTempRes
		out.Thread = rrow.Thread
		out.Profile = entity.Profile{
			Name:      rrow.ProfName,
			NameAlias: rrow.ProfNameAlias,
			Avatar:    rrow.ProfAvatar,
			Institution: entity.Institution{
				Name:  rrow.ProfInstName,
				Alias: rrow.ProfInstAlias,
				Type:  rrow.ProfInstType,
			},
		}

		if err := json.Unmarshal(rrow.AttachmentsJSON, &out.Attachments); err != nil {
			return nil, meta, err
		}
		if err := json.Unmarshal(rrow.TagsJSON, &out.Tags); err != nil {
			return nil, meta, err
		}
		if err := json.Unmarshal(rrow.PartnerTypesJSON, &out.PartnerTypes); err != nil {
			return nil, meta, err
		}
		if err := json.Unmarshal(rrow.InstitutionsJSON, &out.Institutions); err != nil {
			return nil, meta, err
		}

		threads = append(threads, out)
	}
	if err := rows.Err(); err != nil {
		return nil, meta, err
	}
	return
}

func (r *pgsqlThreadRepository) GetDetail(ctx context.Context, request *request.GetDetailThreadReq) (res response.GetDetailThreadTempRes, err error) {
	if request.ID == "" {
		return res, fmt.Errorf("id is required")
	}

	const query = `SELECT
		  -- thread
		  t.id, t.user_id, t.title, t.type, t.description, t.status,
		  t.upvote_number, t.report_number, t.followed_number, t.deadline,
		  t.is_active, t.created_by, t.created_at, t.updated_by, t.updated_at, t.deleted_at,

		  -- profile (1-1)
		  COALESCE(p.name,'')        AS prof_name,
		  COALESCE(p.name_alias,'')  AS prof_name_alias,
		  COALESCE(p.avatar,'') AS prof_avatar,

		  -- institution inside profile
		  COALESCE(i.name,'')  AS prof_inst_name,
		  COALESCE(i.alias,'') AS prof_inst_alias,   -- hapus baris ini kalau kolom alias belum ada di DB
		  COALESCE(i.type,'')  AS prof_inst_type,

		  -- aggregates
		  COALESCE(ja.attachments,'[]'::jsonb)   AS attachments,
		  COALESCE(jtg.tags,'[]'::jsonb)         AS tags,
		  COALESCE(jpt.partner_types,'[]'::jsonb) AS partner_types,
		  COALESCE(ji.institutions,'[]'::jsonb)  AS institutions
		FROM threads t
		LEFT JOIN profiles p     ON p.user_id = t.user_id
		LEFT JOIN institutions i ON i.id = p.institution_id

		-- attachments
		LEFT JOIN LATERAL (
		  SELECT jsonb_agg(
				   jsonb_build_object(
					 'id', ta.id,
					 'file_name', ta.file_name, 'file_url', ta.file_url, 'file_type', ta.file_type,
					 'is_active', ta.is_active, 'created_at', ta.created_at,
					 'updated_at', ta.updated_at
				   ) ORDER BY ta.created_at DESC
				 ) AS attachments
		  FROM thread_attachments ta
		  WHERE ta.thread_id = t.id AND ta.is_active = true
		) ja ON true

		-- tags
		LEFT JOIN LATERAL (
		  SELECT jsonb_agg(
				   jsonb_build_object(
					 'id', tg.id,
					 'name', tg.name,
					 'description', tg.description,
					 'is_active', tt.is_active, 'created_at', tt.created_at,
					 'updated_at', tt.updated_at
				   ) ORDER BY tg.name
				 ) AS tags
		  FROM thread_tags tt
		  JOIN tags tg ON tg.id = tt.tag_id
		  WHERE tt.thread_id = t.id AND tt.is_active = true
		) jtg ON true

		-- partner types
		LEFT JOIN LATERAL (
		  SELECT jsonb_agg(
				   jsonb_build_object(
					 'id', tpt.id,
					 'name', pt.name,
					 'compensation_type', ct.name,
					 'compensation_value', tpt.compensation_value,
					 'compensation_currency', tpt.compensation_currency,
					 'compensation_period', tpt.compensation_period,
					 'compensation_note', tpt.compensation_note,
					 'amount_needed', tpt.amount_needed,
					 'amount_fulfilled', tpt.amount_fulfilled,
					 'is_active', tpt.is_active,
					 'created_at', tpt.created_at,
					 'updated_at', tpt.updated_at
				   ) ORDER BY pt.name
				 ) AS partner_types
		  FROM thread_partner_types tpt
		  JOIN partner_types pt ON pt.id = tpt.partner_type_id
		  LEFT JOIN compensation_types ct ON ct.id = tpt.compensation_type
		  WHERE tpt.thread_id = t.id AND tpt.is_active = true
		) jpt ON true

		-- institutions on thread
		LEFT JOIN LATERAL (
		  SELECT jsonb_agg(
				   jsonb_build_object(
					 'id', i2.id,
					 'name', i2.name,
					 'alias', i2.alias,
					 'type', i2.type,
					 'is_active', ti.is_active, 'created_at', ti.created_at,
					 'updated_at', ti.updated_at
				   ) ORDER BY i2.name
				 ) AS institutions
		  FROM thread_institutions ti
		  JOIN institutions i2 ON i2.id = ti.institution_id
		  WHERE ti.thread_id = t.id AND ti.is_active = true
		) ji ON true
	
	WHERE t.id = $1
	LIMIT 1
	`

	// struct holder untuk scan
	type rowStruct struct {
		entity.Thread

		ProfName      string
		ProfNameAlias string
		ProfAvatar    string
		ProfInstName  string
		ProfInstAlias string
		ProfInstType  string

		AttachmentsJSON  []byte
		TagsJSON         []byte
		PartnerTypesJSON []byte
		InstitutionsJSON []byte
	}

	var row rowStruct

	// nullable holders
	var (
		deadlineNT  sql.NullTime
		updatedByNS sql.NullString
		updatedAtNT sql.NullInt64
		deletedAtNT sql.NullInt64
	)

	sc := r.db.QueryRowContext(ctx, query, request.ID)
	if err = sc.Scan(
		&row.ID, &row.UserID, &row.Title, pq.Array(&row.Type), &row.Description, &row.Status,
		&row.UpvoteNumber, &row.ReportNumber, &row.FollowedNumber, &deadlineNT,
		&row.IsActive, &row.CreatedBy, &row.CreatedAt, &updatedByNS, &updatedAtNT, &deletedAtNT,
		&row.ProfName, &row.ProfNameAlias, &row.ProfAvatar,
		&row.ProfInstName, &row.ProfInstAlias, &row.ProfInstType,
		&row.AttachmentsJSON, &row.TagsJSON, &row.PartnerTypesJSON, &row.InstitutionsJSON,
	); err != nil {
		// preferred: propagate sql.ErrNoRows; kalau mau custom NotFound, map di layer di atas.
		return res, err
	}

	// map nullable → entity.Thread
	if deadlineNT.Valid {
		row.Deadline = &deadlineNT.Time
	}
	if updatedByNS.Valid {
		row.UpdatedBy = updatedByNS.String
	}
	if updatedAtNT.Valid {
		row.UpdatedAt = updatedAtNT.Int64
	}
	if deletedAtNT.Valid {
		row.DeletedAt = deletedAtNT.Int64
	}

	// build response
	res.Thread = row.Thread
	res.Profile = entity.Profile{
		Name:      row.ProfName,
		NameAlias: row.ProfNameAlias,
		Avatar:    row.ProfAvatar,
		Institution: entity.Institution{
			Name:  row.ProfInstName,
			Alias: row.ProfInstAlias,
			Type:  row.ProfInstType,
		},
	}

	if err := json.Unmarshal(row.AttachmentsJSON, &res.Attachments); err != nil {
		return res, err
	}
	if err := json.Unmarshal(row.TagsJSON, &res.Tags); err != nil {
		return res, err
	}
	if err := json.Unmarshal(row.PartnerTypesJSON, &res.PartnerTypes); err != nil {
		return res, err
	}
	if err := json.Unmarshal(row.InstitutionsJSON, &res.Institutions); err != nil {
		return res, err
	}

	return res, nil
}
