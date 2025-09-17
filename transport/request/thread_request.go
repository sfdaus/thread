package request

import (
	"mime/multipart"

	validation "github.com/go-ozzo/ozzo-validation"
)

// CreateThreadReq represent create request body
type PartnerTypeReq struct {
	PartnerTypeID        string  `json:"partner_type_id" validate:"required"`
	CompensationType     string  `json:"compensation_type" validate:"required"`
	CompensationValue    float64 `json:"compensation_value" validate:"required"`
	CompensationCurrency string  `json:"compensation_currency" validate:"required"`
	CompensationPeriod   string  `json:"compensation_period" validate:"omitempty"`
	CompensationNote     string  `json:"compensation_note"`
	AmountNeeded         int64   `json:"amount_needed"`
}

type CreateThreadReq struct {
	UserID       string
	Title        string                  `form:"title"`
	Type         []string                `form:"type"`
	Description  string                  `form:"description"`
	Status       string                  `form:"status"`
	Deadline     string                  `form:"deadline"`
	Attachments  []*multipart.FileHeader `form:"attachments"`
	Tags         []string                `form:"tags"`
	Institutions []string                `form:"institutions"`

	PartnerTypeJSON string           `form:"partner_types"`
	PartnerTypes    []PartnerTypeReq `json:"-"`
}

func (request CreateThreadReq) Validate() error {
	return validation.ValidateStruct(
		&request,
		validation.Field(&request.Title, validation.Required),
		validation.Field(&request.Type, validation.Required),
		validation.Field(&request.Description, validation.Required),
		validation.Field(&request.Status, validation.Required),
		validation.Field(&request.Tags,
			validation.Required,
			validation.Length(1, 5).Error("tag must between 1 and 5."),
		),
		validation.Field(&request.UserID, validation.Required),
	)
}

// Update request body
type PartnerTypeUpdateReq struct {
	ID                   *string  `json:"id"`
	PartnerTypeID        *string  `json:"partner_type_id"`
	CompensationType     *string  `json:"compensation_type"`
	CompensationValue    *float64 `json:"compensation_value"`
	CompensationCurrency *string  `json:"compensation_currency"`
	CompensationPeriod   *string  `json:"compensation_period"`
	CompensationNote     *string  `json:"compensation_note"`
	AmountNeeded         *int64   `json:"amount_needed"`
}
type UpdateThreadReq struct {
	ID                  string `param:"id"`
	UserID              string
	Title               string                  `form:"title"`
	Type                []string                `form:"type[]"`
	Description         string                  `form:"description"`
	Status              string                  `form:"status"`
	Deadline            string                  `form:"deadline"`
	AddedAttachments    []*multipart.FileHeader `form:"added_attachments[]"`
	RemoveAttachmentIDs []string                `form:"remove_attachment_ids[]"`
	AddedTags           []string                `form:"added_tags[]"`
	RemoveTags          []string                `form:"remove_tags[]"`
	AddedInstitutions   []string                `form:"added_institutions[]"`
	RemoveInstitutions  []string                `form:"remove_institutions[]"`

	PartnerTypeJSON string                 `form:"partner_types" validate:"omitempty"`
	PartnerTypes    []PartnerTypeUpdateReq `json:"-"`
}

func (request UpdateThreadReq) Validate() error {
	return validation.ValidateStruct(
		&request,
		validation.Field(&request.ID, validation.Required),
		validation.Field(&request.UserID, validation.Required),
	)
}

// Delete request body
type DeleteThreadReq struct {
	ID     string `param:"id"`
	UserID string
}

func (request DeleteThreadReq) Validate() error {
	return validation.ValidateStruct(
		&request,
		validation.Field(&request.ID, validation.Required),
		validation.Field(&request.UserID, validation.Required),
	)
}

// Get List request body
type GetListThreadReq struct {
	UserID   string
	Title    string `query:"title"`
	Status   string `query:"status"`
	IsActive *bool  `query:"is_active"`
	Time     string `query:"time"`
	PerPage  int64  `query:"per_page"`
	Page     int64  `query:"page"`
}

func (request GetListThreadReq) Validate() error {
	return validation.ValidateStruct(
		&request,
		validation.Field(&request.UserID, validation.Required),
	)
}

// Get Detail request body
type GetDetailThreadReq struct {
	ID     string `param:"id"`
	UserID string
}

func (request GetDetailThreadReq) Validate() error {
	return validation.ValidateStruct(
		&request,
		validation.Field(&request.ID, validation.Required),
		validation.Field(&request.UserID, validation.Required),
	)
}

// Report Thread request body
type ReportThreadReq struct {
	ID          string `param:"id"`
	ReasonID    string `json:"reason_id"`
	Description string `json:"description"`
	UserID      string
}

func (request ReportThreadReq) Validate() error {
	return validation.ValidateStruct(
		&request,
		validation.Field(&request.ID, validation.Required),
		validation.Field(&request.ReasonID, validation.Required),
		validation.Field(&request.UserID, validation.Required),
	)
}

// Like Thread request body
type UpvoteThreadReq struct {
	ID     string `param:"id"`
	UserID string
}

func (request UpvoteThreadReq) Validate() error {
	return validation.ValidateStruct(
		&request,
		validation.Field(&request.ID, validation.Required),
		validation.Field(&request.UserID, validation.Required),
	)
}

// Unvote Thread request body
type UnvoteThreadReq struct {
	ID     string `param:"id"`
	UserID string
}

func (request UnvoteThreadReq) Validate() error {
	return validation.ValidateStruct(
		&request,
		validation.Field(&request.ID, validation.Required),
		validation.Field(&request.UserID, validation.Required),
	)
}

// Share Thread request body
type ShareThreadReq struct {
	ID     string `param:"id"`
	UserID string
}

func (request ShareThreadReq) Validate() error {
	return validation.ValidateStruct(
		&request,
		validation.Field(&request.ID, validation.Required),
		validation.Field(&request.UserID, validation.Required),
	)
}

// Get My Thread request body
type GetMyThreadReq struct {
	UserID   string
	Title    string `query:"title"`
	Status   string `query:"status"`
	IsActive *bool  `query:"is_active"`
	PerPage  int64  `query:"per_page"`
	Page     int64  `query:"page"`
}

func (request GetMyThreadReq) Validate() error {
	return validation.ValidateStruct(
		&request,
		validation.Field(&request.UserID, validation.Required),
	)
}

// Get Shared Thread request body
type GetDetailSharedReq struct {
	Code   string `param:"code"`
	UserID string
}

func (request GetDetailSharedReq) Validate() error {
	return validation.ValidateStruct(
		&request,
		validation.Field(&request.Code, validation.Required),
	)
}

// Follow Thread request body
type FollowThreadReq struct {
	ID     string `param:"id"`
	UserID string
}

func (request FollowThreadReq) Validate() error {
	return validation.ValidateStruct(
		&request,
		validation.Field(&request.ID, validation.Required),
		validation.Field(&request.UserID, validation.Required),
	)
}

// Unfollow Thread request body
type UnfollowThreadReq struct {
	ID     string `param:"id"`
	UserID string
}

func (request UnfollowThreadReq) Validate() error {
	return validation.ValidateStruct(
		&request,
		validation.Field(&request.ID, validation.Required),
		validation.Field(&request.UserID, validation.Required),
	)
}

// Thread Stats request
type ThreadStatsReq struct {
	Filter string `query:"filter"`
	UserID string
}

func (request ThreadStatsReq) Validate() error {
	return validation.ValidateStruct(
		&request,
		validation.Field(&request.UserID, validation.Required),
	)
}

// Thread Activities Follow request
type ThreadFollowActivitiesReq struct {
	PerPage int64 `query:"per_page"`
	Page    int64 `query:"page"`
	UserID  string
}

func (request ThreadFollowActivitiesReq) Validate() error {
	return validation.ValidateStruct(
		&request,
		validation.Field(&request.UserID, validation.Required),
	)
}

// Thread Activities Upvote request
type ThreadUpvoteActivitiesReq struct {
	PerPage int64 `query:"per_page"`
	Page    int64 `query:"page"`
	UserID  string
}

func (request ThreadUpvoteActivitiesReq) Validate() error {
	return validation.ValidateStruct(
		&request,
		validation.Field(&request.UserID, validation.Required),
	)
}

// Thread Activities Comment request
type ThreadCommentActivitiesReq struct {
	PerPage int64 `query:"per_page"`
	Page    int64 `query:"page"`
	UserID  string
}

func (request ThreadCommentActivitiesReq) Validate() error {
	return validation.ValidateStruct(
		&request,
		validation.Field(&request.UserID, validation.Required),
	)
}

// Get My Thread request body
type GetThreadByAuthorReq struct {
	PublicID string `param:"public_id"`
	Title    string `query:"title"`
	Status   string `query:"status"`
	IsActive *bool  `query:"is_active"`
	PerPage  int64  `query:"per_page"`
	Page     int64  `query:"page"`
}

func (request GetThreadByAuthorReq) Validate() error {
	return validation.ValidateStruct(
		&request,
		validation.Field(&request.PublicID, validation.Required),
	)
}
