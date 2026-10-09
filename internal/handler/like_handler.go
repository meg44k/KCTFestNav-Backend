package handler

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v5"

	"github.com/meg44k/KCTFestNav-Backend/internal/usecase"
)

type LikeHandler struct {
	uc *usecase.LikeUsecase
}

func NewLikeHandler(uc *usecase.LikeUsecase) *LikeHandler {
	return &LikeHandler{uc: uc}
}

// 来場者の投票者番号(Next.js のサーバーが cookie から付ける)
func voterOf(c *echo.Context) string { return c.Request().Header.Get("X-Voter") }

// 新しい投票者番号を作る
func (h *LikeHandler) IssueVoter(c *echo.Context) error {
	token, err := h.uc.IssueVoter(c.Request().Context())
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, map[string]string{"voter": token})
}

// その番号で押したブース
func (h *LikeHandler) Mine(c *echo.Context) error {
	ids, err := h.uc.Mine(c.Request().Context(), voterOf(c))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string][]int{"booth_ids": ids})
}

func (h *LikeHandler) Like(c *echo.Context) error {
	id, err := getIDParam(c)
	if err != nil {
		return err
	}
	if err := h.uc.Like(c.Request().Context(), voterOf(c), id); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *LikeHandler) Unlike(c *echo.Context) error {
	id, err := getIDParam(c)
	if err != nil {
		return err
	}
	if err := h.uc.Unlike(c.Request().Context(), voterOf(c), id); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

type likeBucketResponse struct {
	Start time.Time `json:"start"`
	Count int       `json:"count"`
	Burst bool      `json:"burst"`
}

type boothLikesResponse struct {
	BoothID   int                  `json:"booth_id"`
	Name      string               `json:"name"`
	Organizer string               `json:"organizer"`
	Total     int                  `json:"total"`
	Burst     bool                 `json:"burst"`
	Buckets   []likeBucketResponse `json:"buckets"`
}

// クラス展示ごとの数と 10 分ごとの推移(管理者・学生会)
func (h *LikeHandler) Summary(c *echo.Context) error {
	list, err := h.uc.Summary(c.Request().Context())
	if err != nil {
		return err
	}
	res := make([]boothLikesResponse, len(list))
	for i, b := range list {
		buckets := make([]likeBucketResponse, len(b.Buckets))
		for j, x := range b.Buckets {
			buckets[j] = likeBucketResponse{Start: x.Start, Count: x.Count, Burst: x.Burst}
		}
		res[i] = boothLikesResponse{BoothID: b.BoothID, Name: b.Name, Organizer: b.Organizer, Total: b.Total, Burst: b.Burst, Buckets: buckets}
	}
	return c.JSON(http.StatusOK, map[string]any{"booths": res})
}

// そのブースの from 以上 to 未満(RFC 3339)のいいねを消す(管理者・学生会)
func (h *LikeHandler) RemoveRange(c *echo.Context) error {
	id, err := getIDParam(c)
	if err != nil {
		return err
	}
	from, err1 := time.Parse(time.RFC3339, c.QueryParam("from"))
	to, err2 := time.Parse(time.RFC3339, c.QueryParam("to"))
	if err1 != nil || err2 != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "from and to must be RFC 3339")
	}
	n, err := h.uc.RemoveRange(c.Request().Context(), id, from, to)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]int{"removed": n})
}

// 全部消す(管理者)
func (h *LikeHandler) RemoveAll(c *echo.Context) error {
	n, err := h.uc.RemoveAll(c.Request().Context())
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]int{"removed": n})
}
