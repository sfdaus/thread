package entity

type Attachment struct {
	ID        string `json:"id"`
	ThreadID  string `json:"thread_id"`
	FileName  string `json:"file_name"`
	FileUrl   string `json:"file_url"`
	FileType  string `json:"file_type"`
	IsActive  bool   `json:"is_active"`
	CreatedBy string `json:"created_by"`
	CreatedAt int64  `json:"created_at"`
	UpdatedBy string `json:"updated_by"`
	UpdatedAt int64  `json:"updated_at"`
	DeletedAt int64  `json:"deleted_at"`
}

type SecureAttachment struct {
	ID        string `json:"id"`
	ThreadID  string `json:"thread_id"`
	FileName  string `json:"file_name"`
	FileUrl   string `json:"file_url"`
	FileType  string `json:"file_type"`
	IsActive  bool   `json:"is_active"`
	CreatedAt int64  `json:"created_at"`
	UpdatedAt int64  `json:"updated_at"`
}
