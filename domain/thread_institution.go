package domain

type ThreadInstitution struct {
	ID            string `json:"id"`
	ThreadID      string `json:"thread_id"`
	InstitutionID string `json:"institution_id"`
	IsActive      bool   `json:"is_active"`
	CreatedBy     string `json:"created_by"`
	CreatedAt     int64  `json:"created_at"`
	UpdatedBy     string `json:"updated_by"`
	UpdatedAt     int64  `json:"updated_at"`
	DeletedAt     int64  `json:"deleted_at"`
}
