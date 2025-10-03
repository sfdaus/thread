package usecase

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"prakarsa-app/config"
	"prakarsa-app/entity"
	"prakarsa-app/utils"
	"time"

	"github.com/labstack/echo/v4"

	"prakarsa-app/domain"
	"prakarsa-app/repository/redis"
	"prakarsa-app/repository/s3"
	"prakarsa-app/transport/request"
	"prakarsa-app/transport/response"

	"github.com/google/uuid"
)

type threadUsecase struct {
	threadRepo domain.ThreadRepository
	redisRepo  redis.RedisRepository
	s3Repo     s3.S3Repository
	ctxTimeout time.Duration
}

// NewThreadUsecase will create new an threadUsecase object representation of ThreadUsecase interface
func NewThreadUsecase(threadRepo domain.ThreadRepository, redisRepo redis.RedisRepository, s3Repo s3.S3Repository, ctxTimeout time.Duration) *threadUsecase {
	return &threadUsecase{
		threadRepo: threadRepo,
		redisRepo:  redisRepo,
		s3Repo:     s3Repo,
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
	shortID, err := utils.GenerateShortURL(12)
	threadPayload := &entity.Thread{
		ID:             threadUUID,
		UserID:         request.UserID,
		Title:          request.Title,
		Type:           request.Type,
		Description:    request.Description,
		Status:         request.Status,
		UpvoteNumber:   0,
		ReportNumber:   0,
		FollowedNumber: 0,
		ShortID:        shortID,
		Slug:           utils.Slugify(request.Title),
		IsActive:       true,
		CreatedBy:      request.UserID,
		CreatedAt:      time.Now().Unix(),
	}

	if request.Deadline != "" {
		deadline, parseErr := time.Parse("2006-01-02", request.Deadline)
		if parseErr != nil {
			return res, parseErr
		}
		threadPayload.Deadline = &deadline
	}

	// payload attachment
	var threadAttachmentsPayload []*entity.Attachment
	if len(request.Attachments) > 0 {
		for _, attachment := range request.Attachments {
			mimeFromHeader := attachment.Header.Get("Content-Type")

			/*
				Upload File to s3
			*/
			fileURL, err := u.s3Repo.UploadFile(c, attachment, utils.ThreadFileName, utils.ThreadFilePath,
				config.LoadConfig().S3Bucket)

			if err != nil {
				return res, err
			}

			attachmentPayload := &entity.Attachment{
				ID:        uuid.NewString(),
				ThreadID:  threadUUID,
				FileName:  attachment.Filename,
				FileType:  mimeFromHeader,
				FileUrl:   fileURL,
				IsActive:  true,
				CreatedBy: request.UserID,
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
					CreatedBy: request.UserID,
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
					AmountNeeded:         partnerType.AmountNeeded,
					IsActive:             true,
					CreatedAt:            time.Now().Unix(),
					CreatedBy:            request.UserID,
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
					CreatedBy:     request.UserID,
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
		UpdatedBy: request.UserID,
		UserID:    request.UserID,
	}

	if request.Title != "" {
		threadPayload.Title = request.Title
		threadPayload.Slug = utils.Slugify(request.Title)
	}

	if request.Description != "" {
		threadPayload.Description = request.Description
	}

	if request.Status != "" {
		threadPayload.Status = request.Status
	}

	if request.Deadline != "" {
		deadline, parseErr := time.Parse("2006-01-02", request.Deadline)
		if parseErr != nil {
			return parseErr
		}
		threadPayload.Deadline = &deadline
	}

	if len(request.Type) > 0 {
		threadPayload.Type = request.Type
	}

	/*
		Attachments Update
	*/
	var addedThreadAttachmentsPayload []*entity.Attachment

	if len(request.AddedAttachments) > 0 {
		for _, attachment := range request.AddedAttachments {
			mimeFromHeader := attachment.Header.Get("Content-Type")

			/*
				Upload File to s3
			*/
			fileURL, err := u.s3Repo.UploadFile(c, attachment, utils.ThreadFileName, utils.ThreadFilePath,
				config.LoadConfig().S3Bucket)

			if err != nil {
				return err
			}

			attachmentPayload := &entity.Attachment{
				ID:        uuid.NewString(),
				ThreadID:  request.ID,
				FileName:  attachment.Filename,
				FileType:  mimeFromHeader,
				FileUrl:   fileURL,
				IsActive:  true,
				CreatedBy: request.UserID,
				CreatedAt: time.Now().Unix(),
				UpdatedAt: time.Now().Unix(),
				UpdatedBy: request.UserID,
			}

			addedThreadAttachmentsPayload = append(addedThreadAttachmentsPayload, attachmentPayload)
		}
	}

	/*
		Tags Update
	*/
	var addedThreadTagsPayload []*entity.ThreadTag
	if len(request.AddedTags) > 0 {
		for _, tag := range request.AddedTags {
			threadTagPayload := &entity.ThreadTag{
				ID:        uuid.NewString(),
				ThreadID:  request.ID,
				TagID:     tag,
				IsActive:  true,
				CreatedBy: request.UserID,
				CreatedAt: time.Now().Unix(),
				UpdatedAt: time.Now().Unix(),
				UpdatedBy: request.UserID,
			}

			addedThreadTagsPayload = append(addedThreadTagsPayload, threadTagPayload)
		}
	}

	/*
		Institutions Update
	*/
	var addedThreadInstitutionsPayload []*entity.ThreadInstitution
	if len(request.AddedInstitutions) > 0 {
		for _, institution := range request.AddedInstitutions {
			threadInstitutionPayload := &entity.ThreadInstitution{
				ID:            uuid.NewString(),
				ThreadID:      request.ID,
				InstitutionID: institution,
				IsActive:      true,
				CreatedBy:     request.UserID,
				CreatedAt:     time.Now().Unix(),
				UpdatedAt:     time.Now().Unix(),
				UpdatedBy:     request.UserID,
			}

			addedThreadInstitutionsPayload = append(addedThreadInstitutionsPayload, threadInstitutionPayload)
		}
	}

	/*
		Partner Types Update
	*/
	var addedThreadPartnerTypesPayload []*entity.UpdateThreadPartnerType
	var excludeRemovePartnerTypes = []string{}

	if len(request.PartnerTypes) > 0 {
		for _, partnerType := range request.PartnerTypes {
			if partnerType.ID == nil && (partnerType.PartnerTypeID == nil || partnerType.CompensationType == nil ||
				partnerType.CompensationValue == nil || partnerType.CompensationCurrency == nil) {
				return echo.NewHTTPError(http.StatusBadRequest, utils.NewBadRequestError(
					"partner_type_id, compensation_type, compensation_value, and compensation_currency cannot be empty"))
			}

			// Isi payload
			t := true
			threadPartnerTypePayload := &entity.UpdateThreadPartnerType{
				ThreadID:  request.ID,
				IsActive:  &t,
				CreatedBy: request.UserID,
				CreatedAt: time.Now().Unix(),
				UpdatedAt: time.Now().Unix(),
				UpdatedBy: request.UserID,
			}

			if partnerType.ID == nil {
				threadPartnerTypePayload.ID = uuid.NewString()
			} else {
				threadPartnerTypePayload.ID = *partnerType.ID

				excludeRemovePartnerTypes = append(excludeRemovePartnerTypes, *partnerType.ID)
			}

			if partnerType.PartnerTypeID != nil {
				threadPartnerTypePayload.PartnerTypeID = partnerType.PartnerTypeID
			}

			if partnerType.CompensationType != nil {
				threadPartnerTypePayload.CompensationType = partnerType.CompensationType
			}

			if partnerType.CompensationValue != nil {
				threadPartnerTypePayload.CompensationValue = partnerType.CompensationValue
			}

			if partnerType.CompensationCurrency != nil {
				threadPartnerTypePayload.CompensationCurrency = partnerType.CompensationCurrency
			}

			if partnerType.CompensationPeriod != nil {
				threadPartnerTypePayload.CompensationPeriod = partnerType.CompensationPeriod
			}

			if partnerType.CompensationNote != nil {
				threadPartnerTypePayload.CompensationNote = partnerType.CompensationNote
			}

			if partnerType.AmountNeeded != nil {
				threadPartnerTypePayload.AmountNeeded = partnerType.AmountNeeded
			}

			addedThreadPartnerTypesPayload = append(addedThreadPartnerTypesPayload, threadPartnerTypePayload)
		}
	}

	deletedAttachmentsURL, err := u.threadRepo.Update(ctx, threadPayload, addedThreadAttachmentsPayload, request.RemoveAttachmentIDs, addedThreadTagsPayload,
		request.RemoveTags, addedThreadInstitutionsPayload, request.RemoveInstitutions, addedThreadPartnerTypesPayload, excludeRemovePartnerTypes)
	if err != nil {
		return
	}

	// remove file from filestorage
	if len(deletedAttachmentsURL) > 0 {
		var tempDeleteFile []string
		for _, attachment := range deletedAttachmentsURL {
			tempDeleteFile = append(tempDeleteFile, attachment.FileURL)
		}
		_ = u.s3Repo.DeleteBulk(c, config.LoadConfig().S3Bucket, tempDeleteFile)
	}

	return
}

func (u *threadUsecase) Delete(c context.Context, request *request.DeleteThreadReq) (rowsAffected int64, err error) {
	ctx, cancel := context.WithTimeout(c, u.ctxTimeout)
	defer cancel()

	threadPayload := &entity.Thread{
		ID:     request.ID,
		UserID: request.UserID,
	}

	rowsAffected, deletedAttachmentsURL, err := u.threadRepo.Delete(ctx, threadPayload)

	// remove file from filestorage
	if len(deletedAttachmentsURL) > 0 {
		_ = u.s3Repo.DeleteBulk(c, config.LoadConfig().S3Bucket, deletedAttachmentsURL)
	}

	return
}

func (u *threadUsecase) GetList(c context.Context, request *request.GetListThreadReq) (threads []response.GetListThreadRes, meta response.MetaRes, err error) {
	ctx, cancel := context.WithTimeout(c, u.ctxTimeout)
	defer cancel()

	tempThreads, meta, err := u.threadRepo.GetList(ctx, request)

	if len(tempThreads) > 0 {
		for _, temp := range tempThreads {
			threadRes, err := u.mapTempToResGetList(temp, c)
			if err != nil {
				return threads, meta, err
			}
			threads = append(threads, threadRes)
		}
	}

	return
}
func (u *threadUsecase) mapTempToResGetList(tempThread response.GetListThreadTempRes, c context.Context) (res response.GetListThreadRes, err error) {
	res = response.GetListThreadRes{
		ID:             tempThread.Thread.ID,
		Title:          tempThread.Thread.Title,
		Type:           tempThread.Thread.Type,
		Description:    tempThread.Thread.Description,
		Status:         tempThread.Thread.Status,
		UpvoteNumber:   tempThread.Thread.UpvoteNumber,
		ReportNumber:   tempThread.Thread.ReportNumber,
		FollowedNumber: tempThread.Thread.FollowedNumber,
		Deadline:       tempThread.Thread.Deadline,
		Slug:           tempThread.Thread.Slug,
		IsReported:     tempThread.IsReported,
		IsUpvoted:      tempThread.IsUpvoted,
		IsOwner:        tempThread.IsOwner,
		IsFollowing:    tempThread.IsFollowing,
		IsApplying:     tempThread.IsApplying,
		ApprovalStatus: tempThread.ApprovalStatus,
		IsActive:       tempThread.Thread.IsActive,
		CreatedAt:      tempThread.Thread.CreatedAt,
		UpdatedAt:      tempThread.Thread.UpdatedAt,
		CommentCount:   tempThread.CommentCount,
	}
	res.Tags = tempThread.Tags
	res.Institutions = tempThread.Institutions
	res.PartnerTypes = tempThread.PartnerTypes
	res.Profile = tempThread.Profile

	// Thread Attachments - get presigned url
	for _, attachment := range tempThread.Attachments {
		attachment.DownloadUrl, err = u.s3Repo.GetDownloadURL(c, config.LoadConfig().S3Bucket, attachment.FileUrl, false, time.Duration(24*time.Hour))
		if err != nil {
			return res, err
		}
		attachment.FileUrl, err = u.s3Repo.GetPresignedURL(c, config.LoadConfig().S3Bucket, attachment.FileUrl, false, time.Duration(24*time.Hour))
		if err != nil {
			return res, err
		}
		res.Attachments = append(res.Attachments, attachment)
	}

	// Profile Attachments - get presigned url
	res.Profile.Avatar, err = u.s3Repo.GetPresignedURL(c, config.LoadConfig().S3Bucket, tempThread.Profile.Avatar, false, time.Duration(24*time.Hour))
	if err != nil {
		return res, err
	}

	res.CommentCount = tempThread.CommentCount

	return
}

func (u *threadUsecase) GetDetail(c context.Context, request *request.GetDetailThreadReq) (response response.GetDetailThreadRes, err error) {
	ctx, cancel := context.WithTimeout(c, u.ctxTimeout)
	defer cancel()

	tempThread, err := u.threadRepo.GetDetail(ctx, request)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return response, utils.NewNotFoundError("There is no thread with this ID")
		}
	}

	response, err = u.mapTempToResDetail(tempThread, c)
	if err != nil {
		return response, err
	}

	return
}
func (u *threadUsecase) mapTempToResDetail(tempThread response.GetDetailThreadTempRes, c context.Context) (res response.GetDetailThreadRes, err error) {
	res = response.GetDetailThreadRes{
		ID:             tempThread.Thread.ID,
		Title:          tempThread.Thread.Title,
		Type:           tempThread.Thread.Type,
		Description:    tempThread.Thread.Description,
		Status:         tempThread.Thread.Status,
		UpvoteNumber:   tempThread.Thread.UpvoteNumber,
		ReportNumber:   tempThread.Thread.ReportNumber,
		FollowedNumber: tempThread.Thread.FollowedNumber,
		Deadline:       tempThread.Thread.Deadline,
		Slug:           tempThread.Thread.Slug,
		IsReported:     tempThread.IsReported,
		IsUpvoted:      tempThread.IsUpvoted,
		IsOwner:        tempThread.IsOwner,
		IsFollowing:    tempThread.IsFollowing,
		IsApplying:     tempThread.IsApplying,
		ApprovalStatus: tempThread.ApprovalStatus,
		IsActive:       tempThread.Thread.IsActive,
		CreatedAt:      tempThread.Thread.CreatedAt,
		UpdatedAt:      tempThread.Thread.UpdatedAt,
	}
	res.Tags = tempThread.Tags
	res.Institutions = tempThread.Institutions
	res.Profile = tempThread.Profile

	// Thread Attachments - get presigned url
	for _, attachment := range tempThread.Attachments {
		attachment.DownloadUrl, err = u.s3Repo.GetDownloadURL(c, config.LoadConfig().S3Bucket, attachment.FileUrl, false, time.Duration(24*time.Hour))
		if err != nil {
			return res, err
		}
		attachment.FileUrl, err = u.s3Repo.GetPresignedURL(c, config.LoadConfig().S3Bucket, attachment.FileUrl, false, time.Duration(24*time.Hour))
		if err != nil {
			return res, err
		}
		res.Attachments = append(res.Attachments, attachment)
	}

	// Profile Attachments - get presigned url
	res.Profile.Avatar, err = u.s3Repo.GetPresignedURL(c, config.LoadConfig().S3Bucket, tempThread.Profile.Avatar, false, time.Duration(24*time.Hour))
	if err != nil {
		return res, err
	}

	// Partners Attachments - get presigned url
	for _, partner := range tempThread.PartnerTypes {
		if partner.PartnerProfiles != nil {
			for i, profile := range partner.PartnerProfiles {
				partner.PartnerProfiles[i].Avatar, err = u.s3Repo.GetPresignedURL(c, config.LoadConfig().S3Bucket, profile.Avatar, false, time.Duration(24*time.Hour))
				if err != nil {
					return res, err
				}
			}
		}
		res.PartnerTypes = append(res.PartnerTypes, partner)
	}

	return
}

func (u *threadUsecase) ReportThread(c context.Context, request *request.ReportThreadReq) (err error) {
	ctx, cancel := context.WithTimeout(c, u.ctxTimeout)
	defer cancel()

	contentReportPayload := &entity.ContentReport{
		ID:         uuid.NewString(),
		ReporterID: request.UserID,
		ThreadID:   request.ID,
		CommentID:  "",
		ReasonID:   request.ReasonID,
		Status:     "OPEN",
		IsActive:   true,
		CreatedAt:  time.Now().Unix(),
		CreatedBy:  request.UserID,
		UpdatedAt:  time.Now().Unix(),
	}

	err = u.threadRepo.ReportThread(ctx, contentReportPayload)
	return
}

func (u *threadUsecase) UpvoteThread(c context.Context, request *request.UpvoteThreadReq) (err error) {
	ctx, cancel := context.WithTimeout(c, u.ctxTimeout)
	defer cancel()

	contentReportPayload := &entity.ContentUpvote{
		ID:        uuid.NewString(),
		UserID:    request.UserID,
		ThreadID:  request.ID,
		CommentID: "",
		IsActive:  true,
		CreatedAt: time.Now().Unix(),
		CreatedBy: request.UserID,
		UpdatedAt: time.Now().Unix(),
	}

	err = u.threadRepo.UpvoteThread(ctx, contentReportPayload)
	return
}

func (u *threadUsecase) UnvoteThread(c context.Context, request *request.UnvoteThreadReq) (err error) {
	ctx, cancel := context.WithTimeout(c, u.ctxTimeout)
	defer cancel()

	err = u.threadRepo.UnvoteThread(ctx, request)
	return
}

func (u *threadUsecase) ShareThread(c context.Context, request *request.ShareThreadReq) (res response.ShareThreadRes, err error) {
	ctx, cancel := context.WithTimeout(c, u.ctxTimeout)
	defer cancel()

	shortID, err := utils.GenerateShortURL(3)

	var shareEventPayload *entity.ShareEvent
	shareEventPayload = &entity.ShareEvent{
		ID:              uuid.NewString(),
		UserID:          request.UserID,
		ThreadID:        request.ID,
		Action:          "SHARE",
		Counter:         1,
		ShortIdentifier: shortID,
		CreatedAt:       time.Now().Unix(),
		CreatedBy:       request.UserID,
		UpdatedAt:       time.Now().Unix(),
	}

	var t *entity.Thread
	var uniqueID string

	t, uniqueID, err = u.threadRepo.ShareThread(ctx, request, shareEventPayload)

	if err != nil {
		return
	}

	if uniqueID == "" {
		uniqueID = shortID
	}

	res.URL = fmt.Sprintf("%s/t/%s-%s", config.LoadConfig().BaseURLPrakarsa, t.ShortID, uniqueID)

	return
}

func (u *threadUsecase) GetMyThread(c context.Context, request *request.GetMyThreadReq) (threads []response.GetMyThreadRes, meta response.MetaRes, err error) {
	ctx, cancel := context.WithTimeout(c, u.ctxTimeout)
	defer cancel()

	tempThreads, meta, err := u.threadRepo.GetMyThread(ctx, request)

	if len(tempThreads) > 0 {
		for _, temp := range tempThreads {
			threadRes, err := u.mapTempToResGetMyThread(temp, c)
			if err != nil {
				return threads, meta, err
			}
			threads = append(threads, threadRes)
		}
	}

	return
}
func (u *threadUsecase) mapTempToResGetMyThread(tempThread response.GetMyThreadTempRes, c context.Context) (res response.GetMyThreadRes, err error) {
	res = response.GetMyThreadRes{
		ID:             tempThread.Thread.ID,
		Title:          tempThread.Thread.Title,
		Type:           tempThread.Thread.Type,
		Description:    tempThread.Thread.Description,
		Status:         tempThread.Thread.Status,
		UpvoteNumber:   tempThread.Thread.UpvoteNumber,
		ReportNumber:   tempThread.Thread.ReportNumber,
		FollowedNumber: tempThread.Thread.FollowedNumber,
		Deadline:       tempThread.Thread.Deadline,
		Slug:           tempThread.Thread.Slug,
		IsReported:     tempThread.IsReported,
		IsUpvoted:      tempThread.IsUpvoted,
		IsOwner:        tempThread.IsOwner,
		IsFollowing:    tempThread.IsFollowing,
		IsActive:       tempThread.Thread.IsActive,
		CreatedAt:      tempThread.Thread.CreatedAt,
		UpdatedAt:      tempThread.Thread.UpdatedAt,
		CommentCount:   tempThread.CommentCount,
	}
	res.Tags = tempThread.Tags
	res.Institutions = tempThread.Institutions
	res.PartnerTypes = tempThread.PartnerTypes
	res.Profile = tempThread.Profile

	// Thread Attachments - get presigned url
	for _, attachment := range tempThread.Attachments {
		attachment.DownloadUrl, err = u.s3Repo.GetDownloadURL(c, config.LoadConfig().S3Bucket, attachment.FileUrl, false, time.Duration(24*time.Hour))
		if err != nil {
			return res, err
		}
		attachment.FileUrl, err = u.s3Repo.GetPresignedURL(c, config.LoadConfig().S3Bucket, attachment.FileUrl, false, time.Duration(24*time.Hour))
		if err != nil {
			return res, err
		}
		res.Attachments = append(res.Attachments, attachment)
	}

	// Profile Attachments - get presigned url
	res.Profile.Avatar, err = u.s3Repo.GetPresignedURL(c, config.LoadConfig().S3Bucket, tempThread.Profile.Avatar, false, time.Duration(24*time.Hour))
	if err != nil {
		return res, err
	}

	return
}

func (u *threadUsecase) GetDetailShared(c context.Context, request *request.GetDetailSharedReq) (response response.GetDetailSharedRes, err error) {
	ctx, cancel := context.WithTimeout(c, u.ctxTimeout)
	defer cancel()

	tempThread, err := u.threadRepo.GetDetailShared(ctx, request)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return response, utils.NewNotFoundError("There is no thread with this Code")
		}
	}

	response, err = u.mapTempToResDetailShared(tempThread, c)
	if err != nil {
		return response, err
	}

	return
}
func (u *threadUsecase) mapTempToResDetailShared(tempThread response.GetDetailSharedTempRes, c context.Context) (res response.GetDetailSharedRes, err error) {
	res = response.GetDetailSharedRes{
		ID:             tempThread.Thread.ID,
		Title:          tempThread.Thread.Title,
		Type:           tempThread.Thread.Type,
		Description:    tempThread.Thread.Description,
		Status:         tempThread.Thread.Status,
		UpvoteNumber:   tempThread.Thread.UpvoteNumber,
		ReportNumber:   tempThread.Thread.ReportNumber,
		FollowedNumber: tempThread.Thread.FollowedNumber,
		Deadline:       tempThread.Thread.Deadline,
		Slug:           tempThread.Thread.Slug,
		IsReported:     tempThread.IsReported,
		IsUpvoted:      tempThread.IsUpvoted,
		IsOwner:        tempThread.IsOwner,
		IsFollowing:    tempThread.IsFollowing,
		IsActive:       tempThread.Thread.IsActive,
		CreatedAt:      tempThread.Thread.CreatedAt,
		UpdatedAt:      tempThread.Thread.UpdatedAt,
	}
	res.Tags = tempThread.Tags
	res.Institutions = tempThread.Institutions
	res.Profile = tempThread.Profile

	// Thread Attachments - get presigned url
	for _, attachment := range tempThread.Attachments {
		attachment.DownloadUrl, err = u.s3Repo.GetDownloadURL(c, config.LoadConfig().S3Bucket, attachment.FileUrl, false, time.Duration(24*time.Hour))
		if err != nil {
			return res, err
		}
		attachment.FileUrl, err = u.s3Repo.GetPresignedURL(c, config.LoadConfig().S3Bucket, attachment.FileUrl, false, time.Duration(24*time.Hour))
		if err != nil {
			return res, err
		}
		res.Attachments = append(res.Attachments, attachment)
	}

	// Profile Attachments - get presigned url
	res.Profile.Avatar, err = u.s3Repo.GetPresignedURL(c, config.LoadConfig().S3Bucket, tempThread.Profile.Avatar, false, time.Duration(24*time.Hour))
	if err != nil {
		return res, err
	}

	// Partners Attachments - get presigned url
	for _, partner := range tempThread.PartnerTypes {
		if partner.PartnerProfiles != nil {
			for i, profile := range partner.PartnerProfiles {
				partner.PartnerProfiles[i].Avatar, err = u.s3Repo.GetPresignedURL(c, config.LoadConfig().S3Bucket, profile.Avatar, false, time.Duration(24*time.Hour))
				if err != nil {
					return res, err
				}
			}
		}
		res.PartnerTypes = append(res.PartnerTypes, partner)
	}

	return
}

func (u *threadUsecase) FollowThread(c context.Context, request *request.FollowThreadReq) (err error) {
	ctx, cancel := context.WithTimeout(c, u.ctxTimeout)
	defer cancel()

	followThreadPayload := &entity.ThreadFollow{
		ID:        uuid.NewString(),
		UserID:    request.UserID,
		ThreadID:  request.ID,
		IsActive:  true,
		CreatedAt: time.Now().Unix(),
		CreatedBy: request.UserID,
		UpdatedAt: time.Now().Unix(),
	}

	err = u.threadRepo.FollowThread(ctx, followThreadPayload)

	return
}

func (u *threadUsecase) UnfollowThread(c context.Context, request *request.UnfollowThreadReq) (err error) {
	ctx, cancel := context.WithTimeout(c, u.ctxTimeout)
	defer cancel()

	err = u.threadRepo.UnfollowThread(ctx, request)

	return
}

func (u *threadUsecase) ThreadStats(c context.Context, request *request.ThreadStatsReq) (res response.ThreadStatsRes, err error) {
	ctx, cancel := context.WithTimeout(c, u.ctxTimeout)
	defer cancel()

	res, err = u.threadRepo.ThreadStats(ctx, request)

	return
}

func (u *threadUsecase) ThreadFollowActivities(c context.Context, request *request.ThreadFollowActivitiesReq) (res []response.ThreadFollowActivitiesRes, meta response.MetaRes, err error) {
	ctx, cancel := context.WithTimeout(c, u.ctxTimeout)
	defer cancel()

	res, meta, err = u.threadRepo.ThreadFollowActivities(ctx, request)

	for i, r := range res {
		// Profile Attachments - get presigned url
		res[i].Profile.Avatar, err = u.s3Repo.GetPresignedURL(c, config.LoadConfig().S3Bucket, r.Profile.Avatar, false, time.Duration(24*time.Hour))
		if err != nil {
			return res, meta, err
		}

		// Thread Attachments - get presigned url
		for i, attachment := range r.Attachments {
			r.Attachments[i].DownloadUrl, err = u.s3Repo.GetDownloadURL(c, config.LoadConfig().S3Bucket, attachment.FileUrl, false, time.Duration(24*time.Hour))
			if err != nil {
				return res, meta, err
			}
			r.Attachments[i].FileUrl, err = u.s3Repo.GetPresignedURL(c, config.LoadConfig().S3Bucket, attachment.FileUrl, false, time.Duration(24*time.Hour))
			if err != nil {
				return res, meta, err
			}
		}
	}

	return
}

func (u *threadUsecase) ThreadUpvoteActivities(c context.Context, request *request.ThreadUpvoteActivitiesReq) (res []response.ThreadUpvoteActivitiesRes, meta response.MetaRes, err error) {
	ctx, cancel := context.WithTimeout(c, u.ctxTimeout)
	defer cancel()

	res, meta, err = u.threadRepo.ThreadUpvoteActivities(ctx, request)

	for i, r := range res {
		// Profile Attachments - get presigned url
		res[i].Profile.Avatar, err = u.s3Repo.GetPresignedURL(c, config.LoadConfig().S3Bucket, r.Profile.Avatar, false, time.Duration(24*time.Hour))
		if err != nil {
			return res, meta, err
		}

		// Thread Attachments - get presigned url
		for i, attachment := range r.Attachments {
			r.Attachments[i].DownloadUrl, err = u.s3Repo.GetDownloadURL(c, config.LoadConfig().S3Bucket, attachment.FileUrl, false, time.Duration(24*time.Hour))
			if err != nil {
				return res, meta, err
			}
			r.Attachments[i].FileUrl, err = u.s3Repo.GetPresignedURL(c, config.LoadConfig().S3Bucket, attachment.FileUrl, false, time.Duration(24*time.Hour))
			if err != nil {
				return res, meta, err
			}
		}
	}

	return
}

func (u *threadUsecase) ThreadCommentActivities(c context.Context, request *request.ThreadCommentActivitiesReq) (res []response.ThreadCommentActivitiesRes, meta response.MetaRes, err error) {
	ctx, cancel := context.WithTimeout(c, u.ctxTimeout)
	defer cancel()

	res, meta, err = u.threadRepo.ThreadCommentActivities(ctx, request)

	for i, r := range res {
		// Profile Attachments - get presigned url
		res[i].Profile.Avatar, err = u.s3Repo.GetPresignedURL(c, config.LoadConfig().S3Bucket, r.Profile.Avatar, false, time.Duration(24*time.Hour))
		if err != nil {
			return res, meta, err
		}

		// Thread Attachments - get presigned url
		for i, attachment := range r.Attachments {
			r.Attachments[i].DownloadUrl, err = u.s3Repo.GetDownloadURL(c, config.LoadConfig().S3Bucket, attachment.FileUrl, false, time.Duration(24*time.Hour))
			if err != nil {
				return res, meta, err
			}
			r.Attachments[i].FileUrl, err = u.s3Repo.GetPresignedURL(c, config.LoadConfig().S3Bucket, attachment.FileUrl, false, time.Duration(24*time.Hour))
			if err != nil {
				return res, meta, err
			}
		}
	}

	return
}

func (u *threadUsecase) GetThreadByAuthor(c context.Context, request *request.GetThreadByAuthorReq) (threads []response.GetThreadByAuthorRes, meta response.MetaRes, err error) {
	ctx, cancel := context.WithTimeout(c, u.ctxTimeout)
	defer cancel()

	threads, meta, err = u.threadRepo.GetThreadByAuthor(ctx, request)

	for i, r := range threads {
		// Profile Attachments - get presigned url
		threads[i].Profile.Avatar, err = u.s3Repo.GetPresignedURL(c, config.LoadConfig().S3Bucket, r.Profile.Avatar, false, time.Duration(24*time.Hour))
		if err != nil {
			return threads, meta, err
		}

		// Thread Attachments - get presigned url
		for i, attachment := range r.Attachments {
			r.Attachments[i].DownloadUrl, err = u.s3Repo.GetDownloadURL(c, config.LoadConfig().S3Bucket, attachment.FileUrl, false, time.Duration(24*time.Hour))
			if err != nil {
				return threads, meta, err
			}
			r.Attachments[i].FileUrl, err = u.s3Repo.GetPresignedURL(c, config.LoadConfig().S3Bucket, attachment.FileUrl, false, time.Duration(24*time.Hour))
			if err != nil {
				return threads, meta, err
			}
		}
	}

	return
}

func (u *threadUsecase) ThreadStatsByID(c context.Context, request *request.ThreadStatsByIDReq) (res response.ThreadStatsByIDRes, err error) {
	ctx, cancel := context.WithTimeout(c, u.ctxTimeout)
	defer cancel()

	res, err = u.threadRepo.ThreadStatsByID(ctx, request)

	return
}
