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
	apiV1.POST("/threads", handler.Create)
	apiV1.PATCH("/threads/:id", handler.Update)
	apiV1.DELETE("/threads/:id", handler.Delete)
	apiV1.GET("/threads", handler.GetList)
	apiV1.GET("/threads/:id", handler.GetDetail)
	apiV1.GET("/threads/share-url/:id", handler.ShareThread)
	// TODO : apiV1.GET("/threads/my-threads", handler.ShareThread)
	// TODO : apiV1.GET("/threads/detail-shared/:id", handler.GetDetailShared)
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
		return c.JSON(utils.ParseHttpError(err))
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
		return c.JSON(utils.ParseHttpError(err))
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
		return c.JSON(utils.ParseHttpError(err))
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
		return c.JSON(utils.ParseHttpError(err))
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
		return c.JSON(utils.ParseHttpError(err))
	} else {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"message": "Thread upvote submitted",
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
		return c.JSON(utils.ParseHttpError(err))
	} else {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"message": "Success",
			"data":    res,
		})
	}
}
