package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/meg44k/KCTFestNav-Backend/internal/domain"
	"github.com/meg44k/KCTFestNav-Backend/internal/usecase"
)

type StageUsecase interface {
	GetSchedule(ctx context.Context) ([]*domain.StageSection, error)
	CreateSection(ctx context.Context, in usecase.SectionInput) (int, error)
	UpdateSection(ctx context.Context, id int, in usecase.SectionInput) error
	DeleteSection(ctx context.Context, id int) error
	CreateBlock(ctx context.Context, sectionID int, start, end time.Time) (int, error)
	UpdateBlock(ctx context.Context, id int, start, end time.Time) error
	DeleteBlock(ctx context.Context, id int) error
	CreatePerformer(ctx context.Context, blockID int, in usecase.PerformerInput) (int, error)
	UpdatePerformer(ctx context.Context, id int, in usecase.PerformerInput) error
	DeletePerformer(ctx context.Context, id int) error
	MovePerformer(ctx context.Context, id int, d domain.MoveDirection) error
	AdvanceBlock(ctx context.Context, id int) (*domain.StageBlock, error)
	RewindBlock(ctx context.Context, id int) (*domain.StageBlock, error)
}

type StageHandler struct {
	uc  StageUsecase
	now func() time.Time // 「演奏中」の判定に使う時計。テストで差し替える
}

func NewStageHandler(uc StageUsecase, now func() time.Time) *StageHandler {
	return &StageHandler{uc: uc, now: now}
}

// 来場者もフロントも日本時間で扱うので、時刻は +09:00 で返す
var jst = time.FixedZone("JST", 9*3600)

type PerformerResponse struct {
	ID           int    `json:"id"`
	BlockID      int    `json:"block_id"`
	Name         string `json:"name"`
	Detail       string `json:"detail"`
	ThumbnailURL string `json:"thumbnail_url"`
	PerformOrder int    `json:"perform_order"`
}

type StageBlockResponse struct {
	ID           int                 `json:"id"`
	SectionID    int                 `json:"section_id"`
	StartTime    time.Time           `json:"start_time"`
	EndTime      time.Time           `json:"end_time"`
	CurrentOrder int                 `json:"current_order"`
	NowPlaying   bool                `json:"now_playing"`
	Performers   []PerformerResponse `json:"performers"`
}

type StageSectionResponse struct {
	ID        int                  `json:"id"`
	Name      string               `json:"name"`
	Location  string               `json:"location"`
	SortOrder int                  `json:"sort_order"`
	Blocks    []StageBlockResponse `json:"blocks"`
}

type StageResponse struct {
	Sections []StageSectionResponse `json:"sections"`
}

type CreatedResponse struct {
	ID int `json:"id"`
}

func toBlockResponse(b *domain.StageBlock, now time.Time) StageBlockResponse {
	performers := make([]PerformerResponse, len(b.Performers))
	for i, p := range b.Performers {
		performers[i] = PerformerResponse{
			ID:           p.ID,
			BlockID:      p.BlockID,
			Name:         p.Name,
			Detail:       p.Detail,
			ThumbnailURL: p.ThumbnailURL,
			PerformOrder: p.PerformOrder,
		}
	}
	return StageBlockResponse{
		ID:           b.ID,
		SectionID:    b.SectionID,
		StartTime:    b.StartTime.In(jst),
		EndTime:      b.EndTime.In(jst),
		CurrentOrder: b.CurrentOrder,
		NowPlaying:   b.NowPlaying(now),
		Performers:   performers,
	}
}

func (h *StageHandler) GetSchedule(c *echo.Context) error {
	sections, err := h.uc.GetSchedule(c.Request().Context())
	if err != nil {
		return err
	}
	now := h.now()
	res := StageResponse{Sections: make([]StageSectionResponse, len(sections))}
	for i, s := range sections {
		blocks := make([]StageBlockResponse, len(s.Blocks))
		for j, b := range s.Blocks {
			blocks[j] = toBlockResponse(b, now)
		}
		res.Sections[i] = StageSectionResponse{
			ID:        s.ID,
			Name:      s.Name,
			Location:  s.Location,
			SortOrder: s.SortOrder,
			Blocks:    blocks,
		}
	}
	return c.JSON(http.StatusOK, res)
}

type SectionRequest struct {
	Name      string `json:"name"`
	Location  string `json:"location"`
	SortOrder int    `json:"sort_order"`
}

func (r SectionRequest) input() usecase.SectionInput {
	return usecase.SectionInput{Name: r.Name, Location: r.Location, SortOrder: r.SortOrder}
}

type BlockRequest struct {
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}

type PerformerRequest struct {
	Name         string `json:"name"`
	Detail       string `json:"detail"`
	ThumbnailURL string `json:"thumbnail_url"`
}

func (r PerformerRequest) input() usecase.PerformerInput {
	return usecase.PerformerInput{Name: r.Name, Detail: r.Detail, ThumbnailURL: r.ThumbnailURL}
}

type MoveRequest struct {
	Direction domain.MoveDirection `json:"direction"`
}

// URL の :id と JSON の本文を読む。Bind は :id を本文で上書きしうるので id は別に取る
func bindWithID(c *echo.Context, req any) (int, error) {
	id, err := getIDParam(c)
	if err != nil {
		return 0, err
	}
	if req != nil {
		if err := c.Bind(req); err != nil {
			return 0, err
		}
	}
	return id, nil
}

func created(c *echo.Context, id int, err error) error {
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, CreatedResponse{ID: id})
}

func noContent(c *echo.Context, err error) error {
	if err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *StageHandler) CreateSection(c *echo.Context) error {
	var req SectionRequest
	if err := c.Bind(&req); err != nil {
		return err
	}
	id, err := h.uc.CreateSection(c.Request().Context(), req.input())
	return created(c, id, err)
}

func (h *StageHandler) UpdateSection(c *echo.Context) error {
	var req SectionRequest
	id, err := bindWithID(c, &req)
	if err != nil {
		return err
	}
	return noContent(c, h.uc.UpdateSection(c.Request().Context(), id, req.input()))
}

func (h *StageHandler) DeleteSection(c *echo.Context) error {
	id, err := bindWithID(c, nil)
	if err != nil {
		return err
	}
	return noContent(c, h.uc.DeleteSection(c.Request().Context(), id))
}

func (h *StageHandler) CreateBlock(c *echo.Context) error {
	var req BlockRequest
	sectionID, err := bindWithID(c, &req)
	if err != nil {
		return err
	}
	id, err := h.uc.CreateBlock(c.Request().Context(), sectionID, req.StartTime, req.EndTime)
	return created(c, id, err)
}

func (h *StageHandler) UpdateBlock(c *echo.Context) error {
	var req BlockRequest
	id, err := bindWithID(c, &req)
	if err != nil {
		return err
	}
	return noContent(c, h.uc.UpdateBlock(c.Request().Context(), id, req.StartTime, req.EndTime))
}

func (h *StageHandler) DeleteBlock(c *echo.Context) error {
	id, err := bindWithID(c, nil)
	if err != nil {
		return err
	}
	return noContent(c, h.uc.DeleteBlock(c.Request().Context(), id))
}

func (h *StageHandler) CreatePerformer(c *echo.Context) error {
	var req PerformerRequest
	blockID, err := bindWithID(c, &req)
	if err != nil {
		return err
	}
	id, err := h.uc.CreatePerformer(c.Request().Context(), blockID, req.input())
	return created(c, id, err)
}

func (h *StageHandler) UpdatePerformer(c *echo.Context) error {
	var req PerformerRequest
	id, err := bindWithID(c, &req)
	if err != nil {
		return err
	}
	return noContent(c, h.uc.UpdatePerformer(c.Request().Context(), id, req.input()))
}

func (h *StageHandler) DeletePerformer(c *echo.Context) error {
	id, err := bindWithID(c, nil)
	if err != nil {
		return err
	}
	return noContent(c, h.uc.DeletePerformer(c.Request().Context(), id))
}

func (h *StageHandler) MovePerformer(c *echo.Context) error {
	var req MoveRequest
	id, err := bindWithID(c, &req)
	if err != nil {
		return err
	}
	return noContent(c, h.uc.MovePerformer(c.Request().Context(), id, req.Direction))
}

func (h *StageHandler) step(c *echo.Context, fn func(context.Context, int) (*domain.StageBlock, error)) error {
	id, err := bindWithID(c, nil)
	if err != nil {
		return err
	}
	b, err := fn(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, toBlockResponse(b, h.now()))
}

// 演奏中を次の出演者へ進める(学生会・管理者)
func (h *StageHandler) AdvanceBlock(c *echo.Context) error { return h.step(c, h.uc.AdvanceBlock) }

// 押し間違えたときに 1 つ戻す(学生会・管理者)
func (h *StageHandler) RewindBlock(c *echo.Context) error { return h.step(c, h.uc.RewindBlock) }
