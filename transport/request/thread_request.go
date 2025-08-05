package request

import (
	validation "github.com/go-ozzo/ozzo-validation"
	"mime/multipart"
)

// CreateThreadReq represent create todo request body
type CreateThreadReq struct {
	Title        string                  `form:"title"`
	Type         string                  `form:"type"`
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
			validation.Length(1, 0).Error("1 minimal tag required."),
		),
	)
}
