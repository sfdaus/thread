package usecase

import (
	"context"
	"prakarsa-app/entity"
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
	threadPayload := &entity.Thread{
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
		threadPayload.Deadline = &deadline
	}

	// payload attachment
	var threadAttachmentsPayload []*entity.Attachment
	if len(request.Attachments) > 0 {
		for _, attachment := range request.Attachments {
			fileName := strings.Split(attachment.Filename, ".")[0]
			mimeFromHeader := attachment.Header.Get("Content-Type")

			attachmentPayload := &entity.Attachment{
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
	var threadTagsPayload []*entity.ThreadTag
	if len(request.Tags) > 0 {
		for _, tag := range request.Tags {
			threadTagsPayload = append(threadTagsPayload,
				&entity.ThreadTag{
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
	var threadPartnerTypePayload []*entity.ThreadPartnerType

	if len(request.PartnerTypes) > 0 {
		for _, partnerType := range request.PartnerTypes {
			threadPartnerTypePayload = append(threadPartnerTypePayload,
				&entity.ThreadPartnerType{
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
	var threadPartnerInstitutionPayload []*entity.ThreadInstitution

	if len(request.Institutions) > 0 {
		for _, institution := range request.Institutions {
			threadPartnerInstitutionPayload = append(threadPartnerInstitutionPayload,
				&entity.ThreadInstitution{
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

	threadPayload := &entity.Thread{
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
		threadPayload.Deadline = &deadline
	}

	if len(request.Type) > 0 {
		threadPayload.Type = request.Type
	}

	var addedThreadAttachmentsPayload []*entity.Attachment

	if len(request.AddedAttachments) > 0 {
		for _, attachment := range request.AddedAttachments {
			fileName := strings.Split(attachment.Filename, ".")[0]
			mimeFromHeader := attachment.Header.Get("Content-Type")

			attachmentPayload := &entity.Attachment{
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

	threadPayload := &entity.Thread{
		ID: request.ID,
	}

	rowsAffected, err = u.threadRepo.Delete(ctx, threadPayload)
	return
}
func (u *threadUsecase) GetList(c context.Context, request *request.GetListThreadReq) (threads []response.GetListThreadRes, meta response.MetaRes, err error) {
	ctx, cancel := context.WithTimeout(c, u.ctxTimeout)
	defer cancel()

	tempThreads, meta, err := u.threadRepo.GetList(ctx, request)

	if len(tempThreads) > 0 {
		for _, temp := range tempThreads {
			threads = append(threads, mapTempToResGetList(temp))
		}
	}

	return
}
func mapTempToResGetList(tempThread response.GetListThreadTempRes) response.GetListThreadRes {
	res := response.GetListThreadRes{
		ID:             tempThread.Thread.ID,
		Title:          tempThread.Thread.Title,
		Type:           tempThread.Thread.Type,
		Description:    tempThread.Thread.Description,
		Status:         tempThread.Thread.Status,
		UpvoteNumber:   tempThread.Thread.UpvoteNumber,
		ReportNumber:   tempThread.Thread.ReportNumber,
		FollowedNumber: tempThread.Thread.FollowedNumber,
		Deadline:       tempThread.Thread.Deadline,
		IsActive:       tempThread.Thread.IsActive,
		CreatedAt:      tempThread.Thread.CreatedAt,
		UpdatedAt:      tempThread.Thread.UpdatedAt,
	}
	res.Tags = tempThread.Tags
	res.Attachments = tempThread.Attachments
	res.Institutions = tempThread.Institutions
	res.PartnerTypes = tempThread.PartnerTypes
	return res
}
func (u *threadUsecase) GetDetail(c context.Context, request *request.GetDetailThreadReq) (response response.GetDetailThreadRes, err error) {
	ctx, cancel := context.WithTimeout(c, u.ctxTimeout)
	defer cancel()

	tempThread, err := u.threadRepo.GetDetail(ctx, request)

	response = mapTempToResDetail(tempThread)

	return
}
func mapTempToResDetail(tempThread response.GetDetailThreadTempRes) response.GetDetailThreadRes {
	res := response.GetDetailThreadRes{
		ID:             tempThread.Thread.ID,
		Title:          tempThread.Thread.Title,
		Type:           tempThread.Thread.Type,
		Description:    tempThread.Thread.Description,
		Status:         tempThread.Thread.Status,
		UpvoteNumber:   tempThread.Thread.UpvoteNumber,
		ReportNumber:   tempThread.Thread.ReportNumber,
		FollowedNumber: tempThread.Thread.FollowedNumber,
		Deadline:       tempThread.Thread.Deadline,
		IsActive:       tempThread.Thread.IsActive,
		CreatedAt:      tempThread.Thread.CreatedAt,
		UpdatedAt:      tempThread.Thread.UpdatedAt,
	}
	res.Tags = tempThread.Tags
	res.Attachments = tempThread.Attachments
	res.Institutions = tempThread.Institutions
	res.PartnerTypes = tempThread.PartnerTypes
	return res
}
