package entity

type ThreadPartnerType struct {
	ID                   string  `json:"id"`
	ThreadID             string  `json:"thread_id"`
	PartnerTypeID        string  `json:"partner_type_id"`
	CompensationType     string  `json:"compensation_type"`
	CompensationValue    float64 `json:"compensation_value"`
	CompensationCurrency string  `json:"compensation_currency"`
	CompensationPeriod   string  `json:"compensation_period"`
	CompensationNote     string  `json:"compensation_note"`
	AmountNeeded         int64   `json:"amount_needed"`
	AmountFulfilled      int64   `json:"amount_fulfilled"`
	IsActive             bool    `json:"is_active"`
	CreatedBy            string  `json:"created_by"`
	CreatedAt            int64   `json:"created_at"`
	UpdatedBy            string  `json:"updated_by"`
	UpdatedAt            int64   `json:"updated_at"`
	DeletedAt            int64   `json:"deleted_at"`
}

type SecurePartnerType struct {
	ID                   string  `json:"id"`
	Name                 string  `json:"name"`
	CompensationType     string  `json:"compensation_type"`
	CompensationValue    float64 `json:"compensation_value"`
	CompensationCurrency string  `json:"compensation_currency"`
	CompensationPeriod   string  `json:"compensation_period"`
	CompensationNote     string  `json:"compensation_note"`
	AmountNeeded         int64   `json:"amount_needed"`
	AmountFulfilled      int64   `json:"amount_fulfilled"`
	IsActive             bool    `json:"is_active"`
	CreatedAt            int64   `json:"created_at"`
	UpdatedAt            int64   `json:"updated_at"`
}
