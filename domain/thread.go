package domain

import (
	"context"
	"prakarsa-app/transport/request"
	"prakarsa-app/transport/response"
	"time"
)

type Thread struct {
	ID             string    `json:"id"`
	UserID         string    `json:"user_id"`
	Title          string    `json:"title"`
	Type           []string  `json:"type"`
	Description    string    `json:"description"`
	Status         string    `json:"status"`
	UpvoteNumber   int64     `json:"upvote_number"`
	ReportNumber   int64     `json:"report_number"`
	FollowedNumber int64     `json:"followed_number"`
	Deadline       time.Time `json:"deadline"`
	IsActive       bool      `json:"is_active"`
	CreatedBy      string    `json:"created_by"`
	CreatedAt      int64     `json:"created_at"`
	UpdatedBy      string    `json:"updated_by"`
	UpdatedAt      int64     `json:"updated_at"`
	DeletedAt      int64     `json:"deleted_at"`
}

// // ThreadRepository represent the todos repository contract
type ThreadRepository interface {
	Create(ctx context.Context, thread *Thread, attachments []*Attachment, tags []*ThreadTag, institutions []*ThreadInstitution, partnerTypes []*ThreadPartnerType) error
	Update(ctx context.Context, thread *Thread, attachments []*Attachment, removedAttachments []string) error
	Delete(ctx context.Context, thread *Thread) (int64, error)
}

// ThreadUsecase represent the todos usecase contract
type ThreadUsecase interface {
	Create(ctx context.Context, request *request.CreateThreadReq) (response.CreateThreadRes, error)
	Update(ctx context.Context, request *request.UpdateThreadReq) error
	Delete(ctx context.Context, request *request.DeleteThreadReq) (int64, error)
}
