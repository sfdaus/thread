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
}

type CreateThreadReq struct {
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
	)
}

// Update request body
type UpdateThreadReq struct {
	ID                  string                  `param:"id"`
	Title               string                  `form:"title"`
	Type                []string                `form:"type"`
	Description         string                  `form:"description"`
	Status              string                  `form:"status"`
	Deadline            string                  `form:"deadline"`
	AddedAttachments    []*multipart.FileHeader `form:"added_attachments"`
	RemoveAttachmentIDs []string                `form:"remove_attachment_ids"`
	AddedTags           []string                `form:"added_tags"`
	RemoveTags          []string                `form:"remove_tags"`
	AddedInstitutions   []string                `form:"added_institutions"`
	RemoveInstitutions  []string                `form:"remove_institutions"`
}

func (request UpdateThreadReq) Validate() error {
	return validation.ValidateStruct(
		&request,
		validation.Field(&request.ID, validation.Required),
	)
}

// Delete request body
type DeleteThreadReq struct {
	ID string `param:"id"`
}

func (request DeleteThreadReq) Validate() error {
	return validation.ValidateStruct(
		&request,
		validation.Field(&request.ID, validation.Required),
	)
}

// Get List request body
type GetListThreadReq struct {
	Title    string `query:"title"`
	Status   string `query:"status"`
	IsActive *bool  `query:"is_active"`
	PerPage  int64  `query:"per_page"`
	Page     int64  `query:"page"`
}

func (request GetListThreadReq) Validate() error {
	return validation.ValidateStruct(
		&request,
	)
}

// Get Detail request body
type GetDetailThreadReq struct {
	ID string `param:"id"`
}

func (request GetDetailThreadReq) Validate() error {
	return validation.ValidateStruct(
		&request,
		validation.Field(&request.ID, validation.Required),
	)
}
