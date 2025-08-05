package usecase

import (
	"context"
	"strings"
	"time"

	"prakarsa-app/domain"
	"prakarsa-app/repository/redis"
	"prakarsa-app/transport/request"

	"github.com/google/uuid"
)

type threadUsecase struct {
	threadRepo domain.ThreadRepository
	redisRepo  redis.RedisRepository
	ctxTimeout time.Duration
}

// NewThreadUsecase will create new an threadUsecase object representation of ThreadUsecase interface
func NewThreadUsecase(threadRepo domain.ThreadRepository, redisRepo redis.RedisRepository, ctxTimeout time.Duration) *threadUsecase {
	return &threadUsecase{
		threadRepo: threadRepo,
		redisRepo:  redisRepo,
		ctxTimeout: ctxTimeout,
	}
}

func (u *threadUsecase) Create(c context.Context, request *request.CreateThreadReq) (err error) {
	ctx, cancel := context.WithTimeout(c, u.ctxTimeout)
	defer cancel()
	threadUUID := uuid.NewString()

	// Create Payload
	threadPayload := &domain.Thread{
		ID:             threadUUID,
		UserID:         "TODO_user_id",
		Title:          request.Title,
		Type:           request.Type,
		Description:    request.Description,
		Status:         request.Status,
		UpvoteNumber:   0,
		ReportNumber:   0,
		FollowedNumber: 0,
		IsActive:       true,
		CreatedBy:      "TODO_created_by",
		CreatedAt:      time.Now().Unix(),
	}

	var threadAttachmentsPayload []*domain.Attachment
	if len(request.Attachments) > 0 {
		for _, attachment := range request.Attachments {
			fileName := strings.Split(attachment.Filename, ".")[0]
			mimeFromHeader := attachment.Header.Get("Content-Type")

			attachmentPayload := &domain.Attachment{
				ID:        uuid.NewString(),
				ThreadID:  threadUUID,
				FileName:  fileName,
				FileType:  mimeFromHeader,
				FileUrl:   "TODO_file_url",
				IsActive:  true,
				CreatedBy: "TODO_created_by",
				CreatedAt: time.Now().Unix(),
			}

			threadAttachmentsPayload = append(threadAttachmentsPayload, attachmentPayload)
		}
	}

	err = u.threadRepo.Create(ctx, threadPayload, threadAttachmentsPayload)
	return
}
