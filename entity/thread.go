package entity

import (
	"time"
)

type Thread struct {
	ID             string     `json:"id"`
	UserID         string     `json:"user_id"`
	Title          string     `json:"title"`
	Type           []string   `json:"type"`
	Description    string     `json:"description"`
	Status         string     `json:"status"`
	UpvoteNumber   int64      `json:"upvote_number"`
	ReportNumber   int64      `json:"report_number"`
	FollowedNumber int64      `json:"followed_number"`
	Deadline       *time.Time `json:"deadline"`
	ShortID        string     `json:"short_id"`
	Slug           string     `json:"slug"`
	IsActive       bool       `json:"is_active"`
	CreatedBy      string     `json:"created_by"`
	CreatedAt      int64      `json:"created_at"`
	UpdatedBy      string     `json:"updated_by"`
	UpdatedAt      int64      `json:"updated_at"`
	DeletedAt      int64      `json:"deleted_at"`
}
