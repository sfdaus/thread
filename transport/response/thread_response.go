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
	Slug           string                     `json:"slug"`
	IsReported     bool                       `json:"is_reported"`
	IsUpvoted      bool                       `json:"is_upvoted"`
	IsActive       bool                       `json:"is_active"`
	IsOwner        bool                       `json:"is_owner"`
	IsFollowing    bool                       `json:"is_following"`
	CreatedAt      int64                      `json:"created_at"`
	UpdatedAt      int64                      `json:"updated_at"`
	Profile        entity.Profile             `json:"profile"`
	Attachments    []entity.SecureAttachment  `json:"attachments"`
	Tags           []entity.SecureTag         `json:"tags"`
	PartnerTypes   []entity.SecurePartnerType `json:"partner_types"`
	Institutions   []entity.SecureInstitution `json:"institutions"`
}

type GetListThreadTempRes struct {
	Thread       entity.Thread              `json:"thread"`
	IsReported   bool                       `json:"is_reported"`
	IsUpvoted    bool                       `json:"is_upvoted"`
	IsOwner      bool                       `json:"is_owner"`
	IsFollowing  bool                       `json:"is_following"`
	Attachments  []entity.SecureAttachment  `json:"attachments"`
	Tags         []entity.SecureTag         `json:"tags"`
	PartnerTypes []entity.SecurePartnerType `json:"partner_types"`
	Institutions []entity.SecureInstitution `json:"institutions"`
	Profile      entity.Profile             `json:"profile"`
}

type GetDetailThreadRes struct {
	ID             string                     `json:"id"`
	Title          string                     `json:"title"`
	Type           []string                   `json:"type"`
	Description    string                     `json:"description"`
	Status         string                     `json:"status"`
	UpvoteNumber   int64                      `json:"upvote_number"`
	ReportNumber   int64                      `json:"report_number"`
	FollowedNumber int64                      `json:"followed_number"`
	Deadline       *time.Time                 `json:"deadline"`
	Slug           string                     `json:"slug"`
	IsReported     bool                       `json:"is_reported"`
	IsUpvoted      bool                       `json:"is_upvoted"`
	IsActive       bool                       `json:"is_active"`
	IsOwner        bool                       `json:"is_owner"`
	IsFollowing    bool                       `json:"is_following"`
	CreatedAt      int64                      `json:"created_at"`
	UpdatedAt      int64                      `json:"updated_at"`
	Profile        entity.Profile             `json:"profile"`
	Attachments    []entity.SecureAttachment  `json:"attachments"`
	Tags           []entity.SecureTag         `json:"tags"`
	PartnerTypes   []entity.SecurePartnerType `json:"partner_types"`
	Institutions   []entity.SecureInstitution `json:"institutions"`
}

type GetDetailThreadTempRes struct {
	Thread       entity.Thread              `json:"thread"`
	IsReported   bool                       `json:"is_reported"`
	IsUpvoted    bool                       `json:"is_upvoted"`
	IsOwner      bool                       `json:"is_owner"`
	IsFollowing  bool                       `json:"is_following"`
	Attachments  []entity.SecureAttachment  `json:"attachments"`
	Tags         []entity.SecureTag         `json:"tags"`
	PartnerTypes []entity.SecurePartnerType `json:"partner_types"`
	Institutions []entity.SecureInstitution `json:"institutions"`
	Profile      entity.Profile             `json:"profile"`
}

type ShareThreadRes struct {
	URL string `json:"url"`
}

type GetMyThreadRes struct {
	ID             string                     `json:"id"`
	Title          string                     `json:"title"`
	Type           []string                   `json:"type"`
	Description    string                     `json:"description"`
	Status         string                     `json:"status"`
	UpvoteNumber   int64                      `json:"upvote_number"`
	ReportNumber   int64                      `json:"report_number"`
	FollowedNumber int64                      `json:"followed_number"`
	Deadline       *time.Time                 `json:"deadline"`
	Slug           string                     `json:"slug"`
	IsReported     bool                       `json:"is_reported"`
	IsUpvoted      bool                       `json:"is_upvoted"`
	IsActive       bool                       `json:"is_active"`
	IsOwner        bool                       `json:"is_owner"`
	IsFollowing    bool                       `json:"is_following"`
	CreatedAt      int64                      `json:"created_at"`
	UpdatedAt      int64                      `json:"updated_at"`
	Profile        entity.Profile             `json:"profile"`
	Attachments    []entity.SecureAttachment  `json:"attachments"`
	Tags           []entity.SecureTag         `json:"tags"`
	PartnerTypes   []entity.SecurePartnerType `json:"partner_types"`
	Institutions   []entity.SecureInstitution `json:"institutions"`
}

type GetMyThreadTempRes struct {
	Thread       entity.Thread              `json:"thread"`
	IsReported   bool                       `json:"is_reported"`
	IsUpvoted    bool                       `json:"is_upvoted"`
	IsOwner      bool                       `json:"is_owner"`
	IsFollowing  bool                       `json:"is_following"`
	Attachments  []entity.SecureAttachment  `json:"attachments"`
	Tags         []entity.SecureTag         `json:"tags"`
	PartnerTypes []entity.SecurePartnerType `json:"partner_types"`
	Institutions []entity.SecureInstitution `json:"institutions"`
	Profile      entity.Profile             `json:"profile"`
}

type GetDetailSharedRes struct {
	ID             string                     `json:"id"`
	Title          string                     `json:"title"`
	Type           []string                   `json:"type"`
	Description    string                     `json:"description"`
	Status         string                     `json:"status"`
	UpvoteNumber   int64                      `json:"upvote_number"`
	ReportNumber   int64                      `json:"report_number"`
	FollowedNumber int64                      `json:"followed_number"`
	Deadline       *time.Time                 `json:"deadline"`
	Slug           string                     `json:"slug"`
	IsReported     bool                       `json:"is_reported"`
	IsUpvoted      bool                       `json:"is_upvoted"`
	IsActive       bool                       `json:"is_active"`
	IsOwner        bool                       `json:"is_owner"`
	IsFollowing    bool                       `json:"is_following"`
	CreatedAt      int64                      `json:"created_at"`
	UpdatedAt      int64                      `json:"updated_at"`
	Profile        entity.Profile             `json:"profile"`
	Attachments    []entity.SecureAttachment  `json:"attachments"`
	Tags           []entity.SecureTag         `json:"tags"`
	PartnerTypes   []entity.SecurePartnerType `json:"partner_types"`
	Institutions   []entity.SecureInstitution `json:"institutions"`
}

type GetDetailSharedTempRes struct {
	Thread       entity.Thread              `json:"thread"`
	IsReported   bool                       `json:"is_reported"`
	IsUpvoted    bool                       `json:"is_upvoted"`
	IsOwner      bool                       `json:"is_owner"`
	IsFollowing  bool                       `json:"is_following"`
	Attachments  []entity.SecureAttachment  `json:"attachments"`
	Tags         []entity.SecureTag         `json:"tags"`
	PartnerTypes []entity.SecurePartnerType `json:"partner_types"`
	Institutions []entity.SecureInstitution `json:"institutions"`
	Profile      entity.Profile             `json:"profile"`
}
