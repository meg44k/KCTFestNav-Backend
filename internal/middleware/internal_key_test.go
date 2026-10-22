package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
)

func callWithKey(key, header string, send bool) int {
	e := echo.New()
	e.GET("/", func(c *echo.Context) error { return c.NoContent(http.StatusOK) }, InternalKey(key))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if send {
		req.Header.Set("X-Internal-Key", header)
	}
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec.Code
}

func TestInternalKey(t *testing.T) {
	assert.Equal(t, http.StatusOK, callWithKey("k", "k", true))
	assert.Equal(t, http.StatusUnauthorized, callWithKey("k", "x", true), "違う")
	assert.Equal(t, http.StatusUnauthorized, callWithKey("k", "", false), "無い")
	// 設定が空なら、空のヘッダーでも通さない
	assert.Equal(t, http.StatusUnauthorized, callWithKey("", "", true))
	assert.Equal(t, http.StatusUnauthorized, callWithKey("", "", false))
}
