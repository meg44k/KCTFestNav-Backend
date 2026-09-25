package handler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/meg44k/KCTFestNav-Backend/internal/domain"
	"github.com/meg44k/KCTFestNav-Backend/internal/handler"
	"github.com/stretchr/testify/assert"
)

type mockAnnouncementUsecase struct {
	getFn    func(ctx context.Context) (*domain.Announcement, error)
	updateFn func(ctx context.Context, p domain.AnnouncementParams) error
}

func (m *mockAnnouncementUsecase) Get(ctx context.Context) (*domain.Announcement, error) {
	return m.getFn(ctx)
}

func (m *mockAnnouncementUsecase) Update(ctx context.Context, p domain.AnnouncementParams) error {
	return m.updateFn(ctx, p)
}

func TestAnnouncementHandler_Get(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/announcements", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mockUsecase := &mockAnnouncementUsecase{
		getFn: func(ctx context.Context) (*domain.Announcement, error) {
			return &domain.Announcement{Content: "hello world"}, nil
		},
	}
	h := handler.NewAnnouncementHandler(mockUsecase)

	err := h.Get(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.JSONEq(t, `{"content":"hello world"}`, rec.Body.String())
}

func TestAnnouncementHandler_Update(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPut, "/announcements", strings.NewReader(`{"content":"new msg"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	var updatedContent string
	mockUsecase := &mockAnnouncementUsecase{
		updateFn: func(ctx context.Context, p domain.AnnouncementParams) error {
			updatedContent = p.Content
			return nil
		},
	}
	h := handler.NewAnnouncementHandler(mockUsecase)

	err := h.Update(c)
	assert.NoError(t, err)
	assert.Equal(t, "new msg", updatedContent)
	assert.Equal(t, http.StatusNoContent, rec.Code)
}
