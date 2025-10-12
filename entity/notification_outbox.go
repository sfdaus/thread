package entity

type NotificationOutboxInsert struct {
	ID             string
	UserID         string
	Type           string
	ReferenceType  string
	ReferenceID    string
	HeadersJSON    []byte
	Title          string
	Message        string
	ActionURL      *string
	Priority       string
	IdempotencyKey string
	CreatedAt      int64
	UpdatedAt      int64
}
