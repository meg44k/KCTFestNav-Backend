package handler

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v5"

	"github.com/meg44k/KCTFestNav-Backend/internal/domain"
	"github.com/meg44k/KCTFestNav-Backend/internal/usecase"
)

type VisitHandler struct {
	uc *usecase.VisitUsecase
}

func NewVisitHandler(uc *usecase.VisitUsecase) *VisitHandler {
	return &VisitHandler{uc: uc}
}

type recordVisitRequest struct {
	Visitor string `json:"visitor"`
}

// 今日そのブラウザが来たと記録する(Next.js のサーバーが 1 日 1 回だけ呼ぶ)
func (h *VisitHandler) Record(c *echo.Context) error {
	var req recordVisitRequest
	if err := c.Bind(&req); err != nil {
		return domain.ErrInvalidVisitor
	}
	if err := h.uc.Record(c.Request().Context(), req.Visitor); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

// 直近 days 日(既定 7)の日ごとの数と合計(手元の負荷モニターが見る)
func (h *VisitHandler) Count(c *echo.Context) error {
	days := 7
	if s := c.QueryParam("days"); s != "" {
		n, err := strconv.Atoi(s)
		if err != nil {
			return domain.ErrInvalidRange
		}
		days = n
	}
	summary, err := h.uc.Count(c.Request().Context(), days)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, summary)
}
