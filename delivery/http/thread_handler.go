package http

import (
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
	apiV1.POST("/threads", handler.Create)
}

func (h *ThreadHandler) Create(c echo.Context) error {
	ctx := c.Request().Context()
	var req request.CreateThreadReq

	form, err := c.MultipartForm()
	if err != nil {
		return c.JSON(http.StatusBadRequest, "cannot retrieve attachments")
	}

	req.Attachments = form.File["attachments"]

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, utils.NewUnprocessableEntityError(err.Error()))
	}

	if err := req.Validate(); err != nil {
		errVal := err.(validation.Errors)
		return c.JSON(http.StatusBadRequest, utils.NewInvalidInputError(errVal))
	}

	// Handling Tag if it's not empty
	if len(req.Tags) > 0 {
		for _, tag := range req.Tags {
			if tag == "" {
				return c.JSON(http.StatusBadRequest, utils.NewBadRequestError("ln.46 : tags cannot be empty"))
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
				return c.JSON(http.StatusBadRequest, utils.NewBadRequestError("ln.58 : file cannot be empty"))
			}

			if attachment.Size > utils.MaxFileSize {
				return c.JSON(http.StatusBadRequest, utils.NewBadRequestError("ln.62 : file size exceeded the limit"))
			}
		}
	}

	if err := h.ThreadUC.Create(ctx, &req); err != nil {
		return c.JSON(utils.ParseHttpError(err))
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "thread created",
	})
}
