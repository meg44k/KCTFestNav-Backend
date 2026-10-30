package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/meg44k/KCTFestNav-Backend/internal/domain"
	"github.com/meg44k/KCTFestNav-Backend/internal/handler"
	"github.com/meg44k/KCTFestNav-Backend/internal/repository"
	"github.com/meg44k/KCTFestNav-Backend/internal/router"
	"github.com/meg44k/KCTFestNav-Backend/internal/usecase"
)

const visitKey = "visit-e2e-key"

func setupVisitE2E(t *testing.T, now time.Time) *echo.Echo {
	t.Helper()
	// handler のテスト用の論理 DB(他パッケージの FlushDB と干渉しないよう分ける)
	rdb := redis.NewClient(&redis.Options{Addr: "127.0.0.1:6379", DB: 4})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		t.Skipf("テスト用Redisが起動していないためスキップします: %v", err)
	}
	t.Cleanup(func() { rdb.Close() })
	require.NoError(t, rdb.FlushDB(context.Background()).Err())

	uc := usecase.NewVisitUsecase(repository.NewVisitRepository(rdb), func() time.Time { return now })
	e := echo.New()
	e.HTTPErrorHandler = handler.CustomHTTPErrorHandler
	router.VisitRoutes(e, handler.NewVisitHandler(uc), visitKey)
	return e
}

func visitDo(e *echo.Echo, method, path, key, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if key != "" {
		req.Header.Set("X-Internal-Key", key)
	}
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

func TestVisitE2E(t *testing.T) {
	now := time.Date(2026, 10, 31, 1, 0, 0, 0, time.UTC) // 10:00 JST
	e := setupVisitE2E(t, now)
	const a = `{"visitor":"0b7f2c1e-3a4d-4c5e-8f60-7182a3b4c5d6"}`
	const b = `{"visitor":"9a8b7c6d-5e4f-4a3b-9c2d-1e0f2a3b4c5d"}`

	t.Run("合言葉が無い・違うと断る", func(t *testing.T) {
		assert.Equal(t, http.StatusUnauthorized, visitDo(e, http.MethodPost, "/internal/visits", "", a).Code)
		assert.Equal(t, http.StatusUnauthorized, visitDo(e, http.MethodPost, "/internal/visits", "wrong", a).Code)
		assert.Equal(t, http.StatusUnauthorized, visitDo(e, http.MethodGet, "/internal/visits", "", "").Code)
	})

	t.Run("番号の形がおかしいと 400", func(t *testing.T) {
		assert.Equal(t, http.StatusBadRequest, visitDo(e, http.MethodPost, "/internal/visits", visitKey, `{"visitor":"abc"}`).Code)
	})

	t.Run("記録して数える", func(t *testing.T) {
		assert.Equal(t, http.StatusNoContent, visitDo(e, http.MethodPost, "/internal/visits", visitKey, a).Code)
		assert.Equal(t, http.StatusNoContent, visitDo(e, http.MethodPost, "/internal/visits", visitKey, a).Code)
		assert.Equal(t, http.StatusNoContent, visitDo(e, http.MethodPost, "/internal/visits", visitKey, b).Code)

		rec := visitDo(e, http.MethodGet, "/internal/visits?days=2", visitKey, "")
		require.Equal(t, http.StatusOK, rec.Code)
		var got domain.VisitSummary
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
		assert.Equal(t, []domain.VisitDay{{Day: "2026-10-30", Visitors: 0}, {Day: "2026-10-31", Visitors: 2}}, got.Days)
		assert.Equal(t, int64(2), got.Total)
	})

	t.Run("日数が範囲の外だと 400", func(t *testing.T) {
		assert.Equal(t, http.StatusBadRequest, visitDo(e, http.MethodGet, "/internal/visits?days=0", visitKey, "").Code)
		assert.Equal(t, http.StatusBadRequest, visitDo(e, http.MethodGet, "/internal/visits?days=abc", visitKey, "").Code)
	})
}
