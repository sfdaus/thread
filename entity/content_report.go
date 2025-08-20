package entity

type ContentReport struct {
	ID         string `json:"id"`
	ReporterID string `json:"reporter_id"`
	ThreadID   string `json:"thread_id"`
	CommentID  string `json:"comment_id"`
	ReasonID   string `json:"reason_id"`
	Status     string `json:"status"`
	IsActive   bool   `json:"is_active"`
	CreatedBy  string `json:"created_by"`
	CreatedAt  int64  `json:"created_at"`
	UpdatedBy  string `json:"updated_by"`
	UpdatedAt  int64  `json:"updated_at"`
	DeletedAt  int64  `json:"deleted_at"`
}
