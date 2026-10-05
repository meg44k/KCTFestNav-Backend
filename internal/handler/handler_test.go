package handler_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/meg44k/KCTFestNav-Backend/internal/domain"
	"github.com/meg44k/KCTFestNav-Backend/internal/handler"
	"github.com/stretchr/testify/assert"
)

func TestCustomHTTPErrorHandler(t *testing.T) {
	t.Run("空のお知らせは400", func(t *testing.T) {
		e := echo.New()
		rec := httptest.NewRecorder()
		c := e.NewContext(httptest.NewRequest(http.MethodPut, "/manage/announcements", nil), rec)

		handler.CustomHTTPErrorHandler(c, domain.ErrContentRequired)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}
