package handler

import (
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/meg44k/KCTFestNav-Backend/internal/domain"
	"github.com/meg44k/KCTFestNav-Backend/internal/usecase"
)

type AnnouncementHandler struct {
	usecase usecase.AnnouncementUsecase
}

func NewAnnouncementHandler(u usecase.AnnouncementUsecase) *AnnouncementHandler {
	return &AnnouncementHandler{usecase: u}
}

type GetAnnouncementResponse struct {
	Content string `json:"content"`
}

func (h *AnnouncementHandler) Get(c *echo.Context) error {
	announcement, err := h.usecase.Get(c.Request().Context())
	if err != nil {
		return err
	}
	res := GetAnnouncementResponse{
		Content: announcement.Content,
	}

	return c.JSON(http.StatusOK, res)
}

type UpdateAnnouncementRequest struct {
	Content string `json:"content"`
}

func (h *AnnouncementHandler) Update(c *echo.Context) error {
	var req UpdateAnnouncementRequest
	if err := c.Bind(&req); err != nil {
		return err
	}

	if err := h.usecase.Update(c.Request().Context(), domain.AnnouncementParams{
		Content: req.Content,
	}); err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}
