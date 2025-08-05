package http

import (
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/labstack/echo/v4"
	"net/http"
	"prakarsa-app/delivery/middleware"
	"prakarsa-app/domain"
	"prakarsa-app/transport/request"
	"prakarsa-app/utils"
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
	
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, utils.NewUnprocessableEntityError(err.Error()))
	}

	if err := req.Validate(); err != nil {
		errVal := err.(validation.Errors)
		return c.JSON(http.StatusBadRequest, utils.NewInvalidInputError(errVal))
	}

	if err := h.ThreadUC.Create(ctx, &req); err != nil {
		return c.JSON(utils.ParseHttpError(err))
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "thread created",
	})
}
