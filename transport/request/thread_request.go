package request

import validation "github.com/go-ozzo/ozzo-validation"

// CreateThreadReq represent create todo request body
type CreateThreadReq struct {
	Name string `json:"name"`
}

func (request CreateThreadReq) Validate() error {
	return validation.ValidateStruct(
		&request,
		validation.Field(&request.Name, validation.Required),
	)
}
