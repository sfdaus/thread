package domain

import (
	"context"
	"prakarsa-app/transport/request"
	"time"
)

type Thread struct {
	ID          int64     `json:"id"`
	Title       string    `json:"name"`
	Type        string    `json:"type"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// // ThreadRepository represent the todos repository contract
type ThreadRepository interface {
	Create(ctx context.Context, thread *Thread) error
}

// ThreadUsecase represent the todos usecase contract
type ThreadUsecase interface {
	Create(ctx context.Context, request *request.CreateThreadReq) error
}
