package handler

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/meg44k/KCTFestNav-Backend/internal/domain"
)

type BoothUsecase interface {
	GetByID(ctx context.Context, id int) (*domain.Booth, error)
	GetAll(ctx context.Context) ([]*domain.Booth, error)
	Create(
		ctx context.Context,
		name string,
		organizer string,
		detail string,
		congestionStatus domain.CongestionStatus,
		x float32,
		y float32,
		z float32,
	) error
	Delete(ctx context.Context, id int) error
	Update(
		ctx context.Context,
		id int,
		name string,
		organizer string,
		detail string,
		congestionStatus domain.CongestionStatus,
		x float32,
		y float32,
		z float32,
	) error
	UpdateCongestion(
		ctx context.Context,
		id int,
		congestionStatus domain.CongestionStatus,
	) error
}

type BoothHandler struct {
	boothUsecase BoothUsecase
}

func NewBoothHandler(uc BoothUsecase) *BoothHandler {
	return &BoothHandler{
		boothUsecase: uc,
	}
}

type GetBoothResponse struct {
	ID               int                     `json:"id"`
	Name             string                  `json:"name"`
	Organizer        string                  `json:"organizer"`
	Detail           string                  `json:"detail"`
	CongestionStatus domain.CongestionStatus `json:"congestion_status"`
	X                float32                 `json:"x"`
	Y                float32                 `json:"y"`
	Z                float32                 `json:"z"`
}

func (h *BoothHandler) GetByID(c *echo.Context) error {
	id, err := getIDParam(c)
	if err != nil {
		return err
	}
	b, err := h.boothUsecase.GetByID(c.Request().Context(), id)
	if err != nil {
		return err
	}
	res := GetBoothResponse{
		ID:               b.ID,
		Name:             b.Name,
		Organizer:        b.Organizer,
		Detail:           b.Detail,
		CongestionStatus: b.CongestionStatus(),
		X:                b.X,
		Y:                b.Y,
		Z:                b.Z,
	}

	return c.JSON(http.StatusOK, res)
}

type GetAllBoothsResponse struct {
	Booths []GetBoothResponse `json:"booths"`
}

func (h *BoothHandler) GetAll(c *echo.Context) error {
	booths, err := h.boothUsecase.GetAll(c.Request().Context())
	if err != nil {
		return err
	}

	res := make([]GetBoothResponse, len(booths))
	for i, b := range booths {
		res[i] = GetBoothResponse{
			ID:               b.ID,
			Name:             b.Name,
			Organizer:        b.Organizer,
			Detail:           b.Detail,
			CongestionStatus: b.CongestionStatus(),
			X:                b.X,
			Y:                b.Y,
			Z:                b.Z,
		}
	}
	return c.JSON(http.StatusOK, GetAllBoothsResponse{
		Booths: res,
	})
}

type CreateBoothRequest struct {
	Name             string                  `json:"name"`
	Organizer        string                  `json:"organizer"`
	Detail           string                  `json:"detail"`
	CongestionStatus domain.CongestionStatus `json:"congestion_status"`
	X                float32                 `json:"x"`
	Y                float32                 `json:"y"`
	Z                float32                 `json:"z"`
}

func (h *BoothHandler) Create(c *echo.Context) error {
	var req CreateBoothRequest
	if err := c.Bind(&req); err != nil {
		return err
	}
	if err := h.boothUsecase.Create(
		c.Request().Context(),
		req.Name,
		req.Organizer,
		req.Detail,
		req.CongestionStatus,
		req.X,
		req.Y,
		req.Z,
	); err != nil {
		return err
	}
	return c.NoContent(http.StatusCreated)
}

func (h *BoothHandler) Delete(c *echo.Context) error {
	id, err := getIDParam(c)
	if err != nil {
		return err
	}
	if err := h.boothUsecase.Delete(c.Request().Context(), id); err != nil {
		return err
	}
	return nil
}

type UpdateBoothRequest struct {
	Name             string                  `json:"name"`
	Organizer        string                  `json:"organizer"`
	Detail           string                  `json:"detail"`
	CongestionStatus domain.CongestionStatus `json:"congestion_status"`
	X                float32                 `json:"x"`
	Y                float32                 `json:"y"`
	Z                float32                 `json:"z"`
}

func (h *BoothHandler) Update(c *echo.Context) error {
	var req UpdateBoothRequest

	id, err := getIDParam(c)
	if err != nil {
		return err
	}

	if err := c.Bind(&req); err != nil {
		return err

	}
	if err := h.boothUsecase.Update(
		c.Request().Context(),
		id,
		req.Name,
		req.Organizer,
		req.Detail,
		req.CongestionStatus,
		req.X,
		req.Y,
		req.Z,
	); err != nil {
		return err
	}
	return nil
}

type UpdateCongestionStatusRequest struct {
	CongestionStatus domain.CongestionStatus `json:"congestion_status"`
}

func (h *BoothHandler) UpdateCongestion(c *echo.Context) error {
	var req UpdateCongestionStatusRequest

	id, err := getIDParam(c)
	if err != nil {
		return err
	}
	if err := c.Bind(&req); err != nil {
		return err
	}
	if err := h.boothUsecase.UpdateCongestion(c.Request().Context(), id, req.CongestionStatus); err != nil {
		return err
	}
	return nil
}
