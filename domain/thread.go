package domain

import (
	"context"
	"prakarsa-app/entity"
	"prakarsa-app/transport/request"
	"prakarsa-app/transport/response"
)

// // ThreadRepository represent the todos repository contract
type ThreadRepository interface {
	Create(ctx context.Context, thread *entity.Thread, attachments []*entity.Attachment, tags []*entity.ThreadTag,
		institutions []*entity.ThreadInstitution, partnerTypes []*entity.ThreadPartnerType) error
	Update(ctx context.Context, thread *entity.Thread, attachments []*entity.Attachment, removedAttachments []string,
		addedTags []*entity.ThreadTag, removedTags []string, addedInstitutions []*entity.ThreadInstitution, removedInstitutions []string,
		partnerTypes []*entity.UpdateThreadPartnerType, excludeRemovePartnerTypes []string) error
	Delete(ctx context.Context, thread *entity.Thread) (int64, error)
	GetList(ctx context.Context, request *request.GetListThreadReq) ([]response.GetListThreadTempRes, response.MetaRes, error)
	GetDetail(ctx context.Context, request *request.GetDetailThreadReq) (response.GetDetailThreadTempRes, error)
	ReportThread(ctx context.Context, contentReport *entity.ContentReport) error
	UpvoteThread(ctx context.Context, contentUpvote *entity.ContentUpvote) error
	ShareThread(ctx context.Context, request *request.ShareThreadReq, shareEvent *entity.ShareEvent) (*entity.Thread, error)
}

// ThreadUsecase represent the todos usecase contract
type ThreadUsecase interface {
	Create(ctx context.Context, request *request.CreateThreadReq) (response.CreateThreadRes, error)
	Update(ctx context.Context, request *request.UpdateThreadReq) error
	Delete(ctx context.Context, request *request.DeleteThreadReq) (int64, error)
	GetList(ctx context.Context, request *request.GetListThreadReq) ([]response.GetListThreadRes, response.MetaRes, error)
	GetDetail(ctx context.Context, request *request.GetDetailThreadReq) (response.GetDetailThreadRes, error)
	ReportThread(ctx context.Context, request *request.ReportThreadReq) error
	UpvoteThread(ctx context.Context, request *request.UpvoteThreadReq) error
	ShareThread(ctx context.Context, request *request.ShareThreadReq) (response.ShareThreadRes, error)
}
