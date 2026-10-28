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

	// ボットが無い URL を探しに来るので、500(サーバーの失敗)ではなく 404 と返す
	t.Run("無い URL は404", func(t *testing.T) {
		e := echo.New()
		e.HTTPErrorHandler = handler.CustomHTTPErrorHandler
		e.GET("/booths", func(c *echo.Context) error { return c.NoContent(http.StatusOK) })
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/wp/", nil))

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("あるURLでもメソッドが違えば405", func(t *testing.T) {
		e := echo.New()
		e.HTTPErrorHandler = handler.CustomHTTPErrorHandler
		e.POST("/auth/login", func(c *echo.Context) error { return c.NoContent(http.StatusOK) })
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/auth/login", nil))

		assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
	})
}
