package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"prakarsa-app/delivery/middleware"
	"prakarsa-app/domain"
	"prakarsa-app/transport/request"
	"prakarsa-app/utils"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/labstack/echo/v4"
)

type ThreadHandler struct {
	ThreadUC domain.ThreadUsecase
}

// NewThreadHandler will initialize the todo resources endpoint
func NewThreadHandler(e *echo.Echo, middleware *middleware.Middleware, threadUC domain.ThreadUsecase) {
	handler := &ThreadHandler{
		ThreadUC: threadUC,
	}

	apiV1 := e.Group("/api/v1")
	apiV1.POST("/threads/:id/report", handler.ReportThread)
	apiV1.POST("/threads/:id/upvote", handler.UpvoteThread)
	apiV1.DELETE("/threads/:id/upvote", handler.UnvoteThread)
	apiV1.POST("/threads", handler.Create)
	apiV1.PATCH("/threads/:id", handler.Update)
	apiV1.DELETE("/threads/:id", handler.Delete)
	apiV1.GET("/threads", handler.GetList)
	apiV1.GET("/threads/:id", handler.GetDetail)
	apiV1.GET("/threads/share-url/:id", handler.ShareThread)
	apiV1.GET("/threads/mine", handler.GetMyThread)
	apiV1.GET("/threads/shares/:code", handler.GetDetailShared)
	apiV1.POST("/threads/:id/follow", handler.FollowThread)
	apiV1.DELETE("/threads/:id/follow", handler.UnfollowThread)
	apiV1.GET("/threads/mine/stats", handler.ThreadStats)
	apiV1.GET("/threads/activities/follow", handler.ThreadFollowActivities)
}

func (h *ThreadHandler) Create(c echo.Context) error {
	ctx := c.Request().Context()
	var req request.CreateThreadReq

	form, err := c.MultipartForm()
	if err != nil {
		return c.JSON(http.StatusBadRequest, "Failed to retrieve attachments")
	}

	req.Attachments = form.File["attachments"]

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, utils.NewUnprocessableEntityError(err.Error()))
	}

	req.UserID = c.Request().Header.Get("x-user-id")

	if err := req.Validate(); err != nil {
		errVal := err.(validation.Errors)
		return c.JSON(http.StatusBadRequest, utils.NewInvalidInputError(errVal))
	}

	// Handling Tag if it's not empty
	if len(req.Tags) > 0 {
		for _, tag := range req.Tags {
			if tag == "" {
				return c.JSON(http.StatusBadRequest, utils.NewBadRequestError("ln.54 : tags cannot be empty"))
			}
		}
	}

	// Handling Attachment if it's not empty
	if len(req.Attachments) > utils.MaxTotalAttachments {
		errMessage, _ := fmt.Printf("ln.45 : file length cannot more than %d", utils.MaxTotalAttachments)
		return c.JSON(http.StatusBadRequest, utils.NewBadRequestError(errMessage))
	} else if len(req.Attachments) > 0 {
		for _, attachment := range req.Attachments {
			if attachment == nil {
				return c.JSON(http.StatusBadRequest, utils.NewBadRequestError("ln.66 : file cannot be empty"))
			}

			if attachment.Size > utils.MaxFileSize {
				return c.JSON(http.StatusBadRequest, utils.NewBadRequestError("ln.70 : file size exceeded the limit"))
			}
		}
	}

	// parse JSON partner_types kalau ada
	if req.PartnerTypeJSON != "" {
		if err := json.Unmarshal([]byte(req.PartnerTypeJSON), &req.PartnerTypes); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "partner_types must be valid JSON array")
		}
	}

	if res, err := h.ThreadUC.Create(ctx, &req); err != nil {
		return c.JSON(utils.ParseHttpErrorToBasicResponse(err, "Create Thread Failed"))
	} else {
		return c.JSON(http.StatusCreated, map[string]interface{}{
			"message": "Thread successfully created",
			"data": map[string]interface{}{
				"id": res.ID,
			},
		})
	}
}

func (h *ThreadHandler) Update(c echo.Context) error {
	ctx := c.Request().Context()
	var req request.UpdateThreadReq

	form, err := c.MultipartForm()
	if err != nil {
		return c.JSON(http.StatusBadRequest, "Failed to retrieve attachments")
	}

	req.AddedAttachments = form.File["added_attachments"]

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, utils.NewUnprocessableEntityError(err.Error()))
	}

	req.UserID = c.Request().Header.Get("x-user-id")

	if err := req.Validate(); err != nil {
		errVal := err.(validation.Errors)
		return c.JSON(http.StatusBadRequest, utils.NewInvalidInputError(errVal))
	}

	// Handling Tag if it's not empty
	if len(req.AddedTags) > 0 {
		for _, tag := range req.AddedTags {
			if tag == "" {
				return c.JSON(http.StatusBadRequest, utils.NewBadRequestError("ln.108 : tags cannot be empty"))
			}
		}
	}

	// Handling Attachment if it's not empty
	if len(req.AddedAttachments) > utils.MaxTotalAttachments {
		errMessage, _ := fmt.Printf("ln.45 : file length cannot more than %d", utils.MaxTotalAttachments)
		return c.JSON(http.StatusBadRequest, utils.NewBadRequestError(errMessage))
	} else if len(req.AddedAttachments) > 0 {
		for _, attachment := range req.AddedAttachments {
			if attachment == nil {
				return c.JSON(http.StatusBadRequest, utils.NewBadRequestError("ln.66 : file cannot be empty"))
			}

			if attachment.Size > utils.MaxFileSize {
				return c.JSON(http.StatusBadRequest, utils.NewBadRequestError("ln.70 : file size exceeded the limit"))
			}
		}
	}

	if req.PartnerTypeJSON != "" {
		if err := json.Unmarshal([]byte(req.PartnerTypeJSON), &req.PartnerTypes); err != nil {
			return echo.NewHTTPError(400, "partner_types must be valid JSON array")
		}
	}

	if err := h.ThreadUC.Update(ctx, &req); err != nil {
		return c.JSON(utils.ParseHttpErrorToBasicResponse(err, "Update Failed"))
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Thread successfully updated",
	})
}

func (h *ThreadHandler) Delete(c echo.Context) error {
	ctx := c.Request().Context()
	var req request.DeleteThreadReq

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, utils.NewUnprocessableEntityError(err.Error()))
	}

	req.UserID = c.Request().Header.Get("x-user-id")

	if err := req.Validate(); err != nil {
		errVal := err.(validation.Errors)
		return c.JSON(http.StatusBadRequest, utils.NewInvalidInputError(errVal))
	}

	if rowsAffected, err := h.ThreadUC.Delete(ctx, &req); err != nil {
		return c.JSON(utils.ParseHttpErrorToBasicResponse(err, "Delete Thread Failed"))
	} else {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"message": "Thread successfully deleted",
			"data": map[string]int64{
				"rows_affected": rowsAffected,
			},
		})
	}
}

func (h *ThreadHandler) GetList(c echo.Context) error {
	ctx := c.Request().Context()
	var req request.GetListThreadReq

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, utils.NewUnprocessableEntityError(err.Error()))
	}

	req.UserID = c.Request().Header.Get("x-user-id")

	if err := req.Validate(); err != nil {
		errVal := err.(validation.Errors)
		return c.JSON(http.StatusBadRequest, utils.NewInvalidInputError(errVal))
	}

	if res, meta, err := h.ThreadUC.GetList(ctx, &req); err != nil {
		return c.JSON(utils.ParseHttpErrorToBasicResponse(err, "Get List Threads Failed"))
	} else {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"message": "Thread successfully retrieved",
			"data":    res,
			"meta":    meta,
		})
	}
}

func (h *ThreadHandler) GetDetail(c echo.Context) error {
	ctx := c.Request().Context()
	var req request.GetDetailThreadReq

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, utils.NewUnprocessableEntityError(err.Error()))
	}

	req.UserID = c.Request().Header.Get("x-user-id")

	if err := req.Validate(); err != nil {
		errVal := err.(validation.Errors)
		return c.JSON(http.StatusBadRequest, utils.NewInvalidInputError(errVal))
	}

	if res, err := h.ThreadUC.GetDetail(ctx, &req); err != nil {
		return c.JSON(utils.ParseHttpErrorToBasicResponse(err, "Get Detail Failed"))
	} else {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"message": "Thread successfully retrieved",
			"data":    res,
		})
	}
}

func (h *ThreadHandler) ReportThread(c echo.Context) error {
	ctx := c.Request().Context()
	var req request.ReportThreadReq

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, utils.NewUnprocessableEntityError(err.Error()))
	}

	req.UserID = c.Request().Header.Get("x-user-id") // case-insensitive

	if err := req.Validate(); err != nil {
		errVal := err.(validation.Errors)
		return c.JSON(http.StatusBadRequest, utils.NewInvalidInputError(errVal))
	}

	if err := h.ThreadUC.ReportThread(ctx, &req); err != nil {
		return c.JSON(utils.ParseHttpErrorToBasicResponse(err, "Report Thread Failed"))
	} else {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"message": "Thread report submitted",
		})
	}
}

func (h *ThreadHandler) UpvoteThread(c echo.Context) error {
	ctx := c.Request().Context()
	var req = request.UpvoteThreadReq{
		ID:     c.Param("id"),
		UserID: c.Request().Header.Get("x-user-id"),
	}

	if err := req.Validate(); err != nil {
		errVal := err.(validation.Errors)
		return c.JSON(http.StatusBadRequest, utils.NewInvalidInputError(errVal))
	}

	if err := h.ThreadUC.UpvoteThread(ctx, &req); err != nil {
		return c.JSON(utils.ParseHttpErrorToBasicResponse(err, "Upvote Thread Failed"))
	} else {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"message": "Thread upvote submitted",
		})
	}
}

func (h *ThreadHandler) UnvoteThread(c echo.Context) error {
	ctx := c.Request().Context()
	var req = request.UnvoteThreadReq{
		ID:     c.Param("id"),
		UserID: c.Request().Header.Get("x-user-id"),
	}

	if err := req.Validate(); err != nil {
		errVal := err.(validation.Errors)
		return c.JSON(http.StatusBadRequest, utils.NewInvalidInputError(errVal))
	}

	if err := h.ThreadUC.UnvoteThread(ctx, &req); err != nil {
		return c.JSON(utils.ParseHttpErrorToBasicResponse(err, "Unvote Thread Failed"))
	} else {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"message": "Thread unvoted",
		})
	}
}

func (h *ThreadHandler) ShareThread(c echo.Context) error {
	ctx := c.Request().Context()
	var req request.ShareThreadReq

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, utils.NewUnprocessableEntityError(err.Error()))
	}

	req.UserID = c.Request().Header.Get("x-user-id")

	if err := req.Validate(); err != nil {
		errVal := err.(validation.Errors)
		return c.JSON(http.StatusBadRequest, utils.NewInvalidInputError(errVal))
	}

	if res, err := h.ThreadUC.ShareThread(ctx, &req); err != nil {
		return c.JSON(utils.ParseHttpErrorToBasicResponse(err, "Share Thread Failed"))
	} else {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"message": "Success",
			"data":    res,
		})
	}
}

func (h *ThreadHandler) GetMyThread(c echo.Context) error {
	ctx := c.Request().Context()
	var req request.GetMyThreadReq

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, utils.NewUnprocessableEntityError(err.Error()))
	}

	req.UserID = c.Request().Header.Get("x-user-id")

	if err := req.Validate(); err != nil {
		errVal := err.(validation.Errors)
		return c.JSON(http.StatusBadRequest, utils.NewInvalidInputError(errVal))
	}

	if res, meta, err := h.ThreadUC.GetMyThread(ctx, &req); err != nil {
		return c.JSON(utils.ParseHttpErrorToBasicResponse(err, "Get Personal Thread Failed"))
	} else {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"message": "Thread successfully retrieved",
			"data":    res,
			"meta":    meta,
		})
	}
}

func (h *ThreadHandler) GetDetailShared(c echo.Context) error {
	ctx := c.Request().Context()
	var req = request.GetDetailSharedReq{
		Code:   c.Param("code"),
		UserID: c.Request().Header.Get("x-user-id"),
	}

	if err := req.Validate(); err != nil {
		errVal := err.(validation.Errors)
		return c.JSON(http.StatusBadRequest, utils.NewInvalidInputError(errVal))
	}

	if res, err := h.ThreadUC.GetDetailShared(ctx, &req); err != nil {
		return c.JSON(utils.ParseHttpErrorToBasicResponse(err, "Get Shared Thread Failed"))
	} else {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"message": "Thread successfully retrieved",
			"data":    res,
		})
	}
}

func (h *ThreadHandler) FollowThread(c echo.Context) error {
	ctx := c.Request().Context()
	var req = request.FollowThreadReq{
		ID:     c.Param("id"),
		UserID: c.Request().Header.Get("x-user-id"),
	}

	if err := req.Validate(); err != nil {
		errVal := err.(validation.Errors)
		return c.JSON(http.StatusBadRequest, utils.NewInvalidInputError(errVal))
	}

	if err := h.ThreadUC.FollowThread(ctx, &req); err != nil {
		return c.JSON(utils.ParseHttpErrorToBasicResponse(err, "Follow Thread Failed"))
	} else {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"message": "Thread followed",
		})
	}
}

func (h *ThreadHandler) UnfollowThread(c echo.Context) error {
	ctx := c.Request().Context()
	var req = request.UnfollowThreadReq{
		ID:     c.Param("id"),
		UserID: c.Request().Header.Get("x-user-id"),
	}

	if err := req.Validate(); err != nil {
		errVal := err.(validation.Errors)
		return c.JSON(http.StatusBadRequest, utils.NewInvalidInputError(errVal))
	}

	if err := h.ThreadUC.UnfollowThread(ctx, &req); err != nil {
		return c.JSON(utils.ParseHttpErrorToBasicResponse(err, "Unfollow Thread Failed"))
	} else {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"message": "Thread unfollowed",
		})
	}
}

func (h *ThreadHandler) ThreadStats(c echo.Context) error {
	ctx := c.Request().Context()
	var req request.ThreadStatsReq

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, utils.NewUnprocessableEntityError(err.Error()))
	}

	req.UserID = c.Request().Header.Get("x-user-id")

	if err := req.Validate(); err != nil {
		errVal := err.(validation.Errors)
		return c.JSON(http.StatusBadRequest, utils.NewInvalidInputError(errVal))
	}

	if res, err := h.ThreadUC.ThreadStats(ctx, &req); err != nil {
		return c.JSON(utils.ParseHttpErrorToBasicResponse(err, "Get Thread Stats Failed"))
	} else {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"message": "Thread statistics successfully retrieved",
			"data":    res,
		})
	}
}

func (h *ThreadHandler) ThreadFollowActivities(c echo.Context) error {
	ctx := c.Request().Context()
	var req request.ThreadFollowActivitiesReq

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, utils.NewUnprocessableEntityError(err.Error()))
	}

	req.UserID = c.Request().Header.Get("x-user-id")

	if err := req.Validate(); err != nil {
		errVal := err.(validation.Errors)
		return c.JSON(http.StatusBadRequest, utils.NewInvalidInputError(errVal))
	}

	if res, meta, err := h.ThreadUC.ThreadFollowActivities(ctx, &req); err != nil {
		return c.JSON(utils.ParseHttpErrorToBasicResponse(err, "Get Thread Follow Activities Failed"))
	} else {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"message": "Thread follow activities successfully retrieved",
			"data":    res,
			"meta":    meta,
		})
	}
}
