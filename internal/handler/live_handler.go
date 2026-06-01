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
		return err
	}
	return c.NoContent(http.StatusCreated)
}
