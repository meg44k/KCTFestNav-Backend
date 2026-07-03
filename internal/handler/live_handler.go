package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/meg44k/KCTFestNav-Backend/internal/domain"
)

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
	UpdateLiveStatus(ctx context.Context, id int, status domain.LiveStatus) error
	Delete(ctx context.Context, id int) error
	GetByID(ctx context.Context, id int) (*domain.Live, error)
	GetAll(ctx context.Context) ([]*domain.Live, error)
	GetCurrentLive(ctx context.Context) (*domain.Live, error)
}

type LiveHandler struct {
	liveUsecase LiveUsecase
}

func NewLiveHandler(uc LiveUsecase) *LiveHandler {
	return &LiveHandler{
		liveUsecase: uc,
	}
}

type CreateLiveRequest struct {
	Name          string    `json:"name"`
	Detail        string    `json:"detail"`
	ThumbnailURL  string    `json:"thumbnail_url"`
	StartTime     time.Time `json:"start_time"`
	EndTime       time.Time `json:"end_time"`
	SessionNumber int8      `json:"session_number"`
}

func (h *LiveHandler) Create(c *echo.Context) error {
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

func (h *LiveHandler) Update(c *echo.Context) error {
	var req UpdateLiveRequest

	if err := c.Bind(&req); err != nil {
		return err
	}
	// Bindすると、IDがReqParamになった後に、BodyのJSON内のIDが割り当てられるため、IDがJSONの初期値のID=0で上書きされてしまう。
	// そのため、Paramから手動でIDを取り出している。
	id, err := getIDParam(c)
	if err != nil {
		return err
	}

	err = h.liveUsecase.Update(
		c.Request().Context(),
		id,
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

type UpdateLiveStatusRequest struct {
	Status domain.Live `json:"status"`
}

func (h *LiveHandler) UpdateLiveStatus(c *echo.Context) error {
	id, err := getIDParam(c)
	if err != nil {
		return err
	}
	var req UpdateLiveRequest
	if err := c.Bind(&req); err != nil {
		return err
	}
	h.liveUsecase.UpdateLiveStatus(c.Request().Context(), id, req.Status)
	return nil
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

func (h *LiveHandler) GetByID(c *echo.Context) error {
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

type GetAllLivesResponse struct {
	Lives []LiveResponse `json:"lives"`
}

func (h *LiveHandler) GetAll(c *echo.Context) error {
	lives, err := h.liveUsecase.GetAll(c.Request().Context())
	if err != nil {
		return err
	}
	res := make([]LiveResponse, len(lives))
	for i, l := range lives {
		resLive := LiveResponse{
			ID:            l.ID,
			Name:          l.Name,
			Detail:        l.Detail,
			ThumbnailURL:  l.ThumbnailURL,
			StartTime:     l.StartTime,
			EndTime:       l.EndTime,
			SessionNumber: l.SessionNumber(),
			Status:        l.Status(),
		}
		res[i] = resLive
	}

	return c.JSON(http.StatusOK, GetAllLivesResponse{
		Lives: res,
	})
}

func (h *LiveHandler) GetCurrentLive(c *echo.Context) error {
	live, err := h.liveUsecase.GetCurrentLive(c.Request().Context())
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

func (h *LiveHandler) Delete(c *echo.Context) error {
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
