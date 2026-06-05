package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/meg44k/KCTFestNav-Backend/internal/domain"
)

type CreateLiveRequest struct {
	Name          string    `json:"name"`
	Detail        string    `json:"detail"`
	ThumbnailURL  string    `json:"thumbnail_url"`
	StartTime     time.Time `json:"start_time"`
	EndTime       time.Time `json:"end_time"`
	SessionNumber int8      `json:"session_number"`
}

type UpdateLiveRequest struct {
	ID            int               `param:"id"`
	Name          string            `json:"name"`
	Detail        string            `json:"detail"`
	ThumbnailURL  string            `json:"thumbnail_url"`
	StartTime     time.Time         `json:"start_time"`
	EndTime       time.Time         `json:"end_time"`
	SessionNumber int8              `json:"session_number"`
	Status        domain.LiveStatus `json:"status"`
}

type LiveResponse struct {
	ID            int               `json:"id"`
	Name          string            `json:"name"`
	Detail        string            `json:"detail"`
	ThumbnailURL  string            `json:"thumbnail_url"`
	StartTime     time.Time         `json:"start_time"`
	EndTime       time.Time         `json:"end_time"`
	SessionNumber int8              `json:"session_number"`
	Status        domain.LiveStatus `json:"status"`
}

type GetAllLivesResponse struct {
	Lives []LiveResponse `json:"lives"`
}

type LiveUsecase interface {
	Create(
		ctx context.Context,
		name string,
		detail string,
		thumbnailURL string,
		startTime time.Time,
		endTime time.Time,
		sessionNumber int8,
	) error
	Update(
		ctx context.Context,
		id int,
		name string,
		detail string,
		thumbnailURL string,
		startTime time.Time,
		endTime time.Time,
		sessionNumber int8,
		status domain.LiveStatus,
	) error
	Delete(ctx context.Context, id int) error
	GetByID(ctx context.Context, id int) (*domain.Live, error)
	GetAll(ctx context.Context) ([]*domain.Live, error)
}

type liveHandler struct {
	liveUsecase LiveUsecase
}

func NewLiveHandler(uc LiveUsecase) *liveHandler {
	return &liveHandler{
		liveUsecase: uc,
	}
}

func (h *liveHandler) Create(c *echo.Context) error {
	var req CreateLiveRequest
	// リクエストをJSONから型にバインドする
	if err := c.Bind(&req); err != nil {
		return err
	}
	ctx := c.Request().Context()
	err := h.liveUsecase.Create(
		ctx,
		req.Name,
		req.Detail,
		req.ThumbnailURL,
		req.StartTime,
		req.EndTime,
		req.SessionNumber,
	)
	if err != nil {
		return err // エラーはCustomErrorHandlerで振り分けされる
	}
	return c.NoContent(http.StatusCreated)
}

func (h *liveHandler) Update(c *echo.Context) error {
	var req UpdateLiveRequest

	if err := c.Bind(&req); err != nil {
		return err
	}
	ctx := c.Request().Context()
	err := h.liveUsecase.Update(
		ctx,
		req.ID,
		req.Name,
		req.Detail,
		req.ThumbnailURL,
		req.StartTime,
		req.EndTime,
		req.SessionNumber,
		req.Status,
	)
	if err != nil {
		return err // エラーはCustomErrorHandlerで振り分けされる
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *liveHandler) GetByID(c *echo.Context) error {
	id, err := getIDParam(c)
	if err != nil {
		return err
	}
	live, err := h.liveUsecase.GetByID(c.Request().Context(), id)
	if err != nil {
		return err
	}

	res := LiveResponse{
		ID:            live.ID,
		Name:          live.Name,
		Detail:        live.Detail,
		ThumbnailURL:  live.ThumbnailURL,
		StartTime:     live.StartTime,
		EndTime:       live.EndTime,
		SessionNumber: live.SessionNumber(),
		Status:        live.Status(),
	}

	return c.JSON(http.StatusOK, res)
}

func (h *liveHandler) GetAll(c *echo.Context) error {
	lives, err := h.liveUsecase.GetAll(c.Request().Context())
	if err != nil {
		return err
	}
	res := make([]LiveResponse, 0, len(lives)) // cap指定
	for _, live := range lives {
		res = append(res, LiveResponse{
			ID:            live.ID,
			Name:          live.Name,
			Detail:        live.Detail,
			ThumbnailURL:  live.ThumbnailURL,
			StartTime:     live.StartTime,
			EndTime:       live.EndTime,
			SessionNumber: live.SessionNumber(),
			Status:        live.Status(),
		})
	}

	return c.JSON(http.StatusOK, GetAllLivesResponse{
		Lives: res,
	})
}

func (h *liveHandler) Delete(c *echo.Context) error {
	id, err := getIDParam(c)
	if err != nil {
		return err
	}
	err = h.liveUsecase.Delete(c.Request().Context(), id)
	if err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}
