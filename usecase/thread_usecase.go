package usecase

import (
	"context"
	"strings"
	"time"

	"prakarsa-app/domain"
	"prakarsa-app/repository/redis"
	"prakarsa-app/transport/request"
	"prakarsa-app/transport/response"

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

func (u *threadUsecase) Create(c context.Context, request *request.CreateThreadReq) (res response.CreateThreadRes, err error) {
	ctx, cancel := context.WithTimeout(c, u.ctxTimeout)
	defer cancel()
	threadUUID := uuid.NewString()
	res.ID = threadUUID

	/**
		Create Payload
	**/

	// payload thread
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

		IsActive:  true,
		CreatedBy: "TODO_created_by",
		CreatedAt: time.Now().Unix(),
	}

	if request.Deadline != "" {
		deadline, parseErr := time.Parse("02-01-2006", request.Deadline)
		if parseErr != nil {
			return res, parseErr
		}
		threadPayload.Deadline = deadline
	}

	// payload attachment
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

	// payload relation thread tags
	var threadTagsPayload []*domain.ThreadTag
	if len(request.Tags) > 0 {
		for _, tag := range request.Tags {
			threadTagsPayload = append(threadTagsPayload,
				&domain.ThreadTag{
					ID:        uuid.NewString(),
					ThreadID:  threadUUID,
					TagID:     tag,
					IsActive:  true,
					CreatedAt: time.Now().Unix(),
					CreatedBy: "TODO_created_by",
				},
			)
		}
	}

	// payload relation thread partner type
	var threadPartnerTypePayload []*domain.ThreadPartnerType

	if len(request.PartnerTypes) > 0 {
		for _, partnerType := range request.PartnerTypes {
			threadPartnerTypePayload = append(threadPartnerTypePayload,
				&domain.ThreadPartnerType{
					ID:                   uuid.NewString(),
					ThreadID:             threadUUID,
					PartnerTypeID:        partnerType.PartnerTypeID,
					CompensationType:     partnerType.CompensationType,
					CompensationValue:    partnerType.CompensationValue,
					CompensationCurrency: partnerType.CompensationCurrency,
					CompensationPeriod:   partnerType.CompensationPeriod,
					CompensationNote:     partnerType.CompensationNote,
					IsActive:             true,
					CreatedAt:            time.Now().Unix(),
					CreatedBy:            "TODO_created_by",
				},
			)
		}
	}

	// payload relation thread institution
	var threadPartnerInstitutionPayload []*domain.ThreadInstitution

	if len(request.Institutions) > 0 {
		for _, institution := range request.Institutions {
			threadPartnerInstitutionPayload = append(threadPartnerInstitutionPayload,
				&domain.ThreadInstitution{
					ID:            uuid.NewString(),
					ThreadID:      threadUUID,
					InstitutionID: institution,
					IsActive:      true,
					CreatedAt:     time.Now().Unix(),
					CreatedBy:     "TODO_created_by",
				},
			)
		}
	}

	err = u.threadRepo.Create(ctx, threadPayload, threadAttachmentsPayload, threadTagsPayload, threadPartnerInstitutionPayload,
		threadPartnerTypePayload)
	return
}
func (u *threadUsecase) Update(c context.Context, request *request.UpdateThreadReq) (err error) {
	ctx, cancel := context.WithTimeout(c, u.ctxTimeout)
	defer cancel()

	threadPayload := &domain.Thread{
		ID:        request.ID,
		UpdatedAt: time.Now().Unix(),
		UpdatedBy: "TODO_updated_by",
	}

	if request.Title != "" {
		threadPayload.Title = request.Title
	}

	if request.Description != "" {
		threadPayload.Description = request.Description
	}

	if request.Status != "" {
		threadPayload.Status = request.Status
	}

	if request.Deadline != "" {
		deadline, parseErr := time.Parse("02-01-2006", request.Deadline)
		if parseErr != nil {
			return parseErr
		}
		threadPayload.Deadline = deadline
	}

	if len(request.Type) > 0 {
		threadPayload.Type = request.Type
	}

	var addedThreadAttachmentsPayload []*domain.Attachment

	if len(request.AddedAttachments) > 0 {
		for _, attachment := range request.AddedAttachments {
			fileName := strings.Split(attachment.Filename, ".")[0]
			mimeFromHeader := attachment.Header.Get("Content-Type")

			attachmentPayload := &domain.Attachment{
				ID:        uuid.NewString(),
				ThreadID:  request.ID,
				FileName:  fileName,
				FileType:  mimeFromHeader,
				FileUrl:   "TODO_file_url",
				IsActive:  true,
				CreatedBy: "TODO_created_by",
				CreatedAt: time.Now().Unix(),
			}

			addedThreadAttachmentsPayload = append(addedThreadAttachmentsPayload, attachmentPayload)
		}
	}

	err = u.threadRepo.Update(ctx, threadPayload, addedThreadAttachmentsPayload, request.RemoveAttachmentIDs)
	return
}
func (u *threadUsecase) Delete(c context.Context, request *request.DeleteThreadReq) (rowsAffected int64, err error) {
	ctx, cancel := context.WithTimeout(c, u.ctxTimeout)
	defer cancel()

	threadPayload := &domain.Thread{
		ID: request.ID,
	}

	rowsAffected, err = u.threadRepo.Delete(ctx, threadPayload)
	return
}
func (u *threadUsecase) GetList(c context.Context, request *request.GetListThreadReq) (threads []response.GetListThreadRes, err error) {
	return
}
