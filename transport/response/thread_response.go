package response

import (
	"prakarsa-app/entity"
	"time"
)

type CreateThreadRes struct {
	ID string `json:"id"`
}

type GetListThreadRes struct {
	ID             string                     `json:"id"`
	Title          string                     `json:"title"`
	Type           []string                   `json:"type"`
	Description    string                     `json:"description"`
	Status         string                     `json:"status"`
	UpvoteNumber   int64                      `json:"upvote_number"`
	ReportNumber   int64                      `json:"report_number"`
	FollowedNumber int64                      `json:"followed_number"`
	Deadline       *time.Time                 `json:"deadline"`
	IsActive       bool                       `json:"is_active"`
	CreatedAt      int64                      `json:"created_at"`
	UpdatedAt      int64                      `json:"updated_at"`
	Attachments    []entity.SecureAttachment  `json:"attachments"`
	Tags           []entity.SecureTag         `json:"tags"`
	PartnerTypes   []entity.SecurePartnerType `json:"partner_types"`
	Institutions   []entity.SecureInstitution `json:"institutions"`
}

type GetListThreadTempRes struct {
	Thread       entity.Thread              `json:"thread"`
	Attachments  []entity.Attachment        `json:"attachments"`
	Tags         []entity.ThreadTag         `json:"tags"`
	PartnerTypes []entity.ThreadPartnerType `json:"partner_types"`
	Institutions []entity.ThreadInstitution `json:"institutions"`
}
