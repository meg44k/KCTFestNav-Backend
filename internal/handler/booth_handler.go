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
		congestionStatus int8,
		x float32,
		y float32,
		z float32,
	) error
	Delete(ctx context.Context, id int) error
}

type BoothHandler struct {
	boothUsecase BoothUsecase
}

func NewBoothHandler(uc BoothUsecase) *BoothHandler {
	return &BoothHandler{
		boothUsecase: uc,
	}
}

type BoothResponse struct {
	ID               int     `json:"id"`
	Name             string  `json:"name"`
	Organizer        string  `json:"organizer"`
	Detail           string  `json:"detail"`
	CongestionStatus int8    `json:"congestionStatus"`
	X                float32 `json:"x"`
	Y                float32 `json:"y"`
	Z                float32 `json:"z"`
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
	res := BoothResponse{
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
	Booths []BoothResponse `json:"booths"`
}

func (h *BoothHandler) GetAll(c *echo.Context) error {
	booths, err := h.boothUsecase.GetAll(c.Request().Context())
	if err != nil {
		return err
	}

	res := make([]BoothResponse, len(booths))
	for i, b := range booths {
		res[i] = BoothResponse{
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
	Name             string  `json:"name"`
	Organizer        string  `json:"organizer"`
	Detail           string  `json:"detail"`
	CongestionStatus int8    `json:"congestionStatus"`
	X                float32 `json:"x"`
	Y                float32 `json:"y"`
	Z                float32 `json:"z"`
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
