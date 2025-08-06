package request

import (
	"mime/multipart"

	validation "github.com/go-ozzo/ozzo-validation"
)

// CreateThreadReq represent create request body
type CreateThreadReq struct {
	Title        string                  `form:"title"`
	Type         []string                `form:"type"`
	Description  string                  `form:"description"`
	Status       string                  `form:"status"`
	Deadline     string                  `form:"deadline"`
	Attachments  []*multipart.FileHeader `form:"attachments"`
	Tags         []string                `form:"tags"`
	Institutions []string                `form:"institutions"`
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
