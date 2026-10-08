package handler_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/meg44k/KCTFestNav-Backend/internal/auth"
	"github.com/meg44k/KCTFestNav-Backend/internal/domain"
	"github.com/meg44k/KCTFestNav-Backend/internal/handler"
	"github.com/meg44k/KCTFestNav-Backend/internal/middleware"
	"github.com/meg44k/KCTFestNav-Backend/internal/repository"
	"github.com/meg44k/KCTFestNav-Backend/internal/router"
	"github.com/meg44k/KCTFestNav-Backend/internal/usecase"
)

// 本物のルーターに、時計を差し替えたステージの handler をつなぐ
func setupStageE2E(t *testing.T, now *time.Time) (*echo.Echo, []byte) {
	t.Helper()
	db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/kctfest_test_handler?parseTime=true")
	if err != nil {
		t.Fatalf("DBの初期化エラー: %v", err)
	}
	if err := db.Ping(); err != nil {
		t.Skipf("テスト用DBが起動していないためスキップします: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	for _, table := range []string{"performers", "stage_blocks", "stage_sections"} {
		_, err := db.Exec("DELETE FROM " + table)
		require.NoError(t, err)
	}

	secret := []byte("e2e-test-secret")
	h := handler.NewStageHandler(usecase.NewStageUsecase(repository.NewStageRepository(db)), func() time.Time { return *now })
	e := echo.New()
	e.HTTPErrorHandler = handler.CustomHTTPErrorHandler
	router.StageRoutes(e, e.Group("/manage", middleware.JWTAuth(secret)), h)
	return e, secret
}

type caller struct {
	t     *testing.T
	e     *echo.Echo
	token string
}

func (c caller) do(method, path string, body any) *httptest.ResponseRecorder {
	c.t.Helper()
	var buf bytes.Buffer
	if body != nil {
		require.NoError(c.t, json.NewEncoder(&buf).Encode(body))
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	rec := httptest.NewRecorder()
	c.e.ServeHTTP(rec, req)
	return rec
}

func (c caller) createdID(method, path string, body any) int {
	c.t.Helper()
	rec := c.do(method, path, body)
	require.Equal(c.t, http.StatusCreated, rec.Code, rec.Body.String())
	var res struct {
		ID int `json:"id"`
	}
	require.NoError(c.t, json.Unmarshal(rec.Body.Bytes(), &res))
	require.NotZero(c.t, res.ID)
	return res.ID
}

func tokenFor(t *testing.T, role domain.Role, secret []byte) string {
	tok, err := auth.GenerateToken(uuid.New(), role, 0, secret)
	require.NoError(t, err)
	return tok
}

func TestStageE2E(t *testing.T) {
	jst := time.FixedZone("JST", 9*3600)
	start := time.Date(2026, 10, 31, 13, 0, 0, 0, jst)
	now := start.Add(10 * time.Minute)
	e, secret := setupStageE2E(t, &now)

	admin := caller{t, e, tokenFor(t, domain.RoleAdmin, secret)}
	gakuseikai := caller{t, e, tokenFor(t, domain.RoleGakuseikai, secret)}
	student := caller{t, e, tokenFor(t, domain.RoleStudent, secret)}
	public := caller{t, e, ""}

	sid := admin.createdID(http.MethodPost, "/manage/stage/sections",
		map[string]any{"name": "Live1", "location": "第一体育館", "sort_order": 1})
	bid := admin.createdID(http.MethodPost, fmt.Sprintf("/manage/stage/sections/%d/blocks", sid),
		map[string]any{"start_time": "2026-10-31T13:00:00+09:00", "end_time": "2026-10-31T13:50:00+09:00"})
	var pids []int
	for _, n := range []string{"バンドA", "バンドB", "バンドC"} {
		pids = append(pids, admin.createdID(http.MethodPost, fmt.Sprintf("/manage/stage/blocks/%d/performers", bid),
			map[string]any{"name": n, "detail": n + "です", "thumbnail_url": ""}))
	}

	t.Run("学生会が次のバンドへ進めると、更新後のブロックが返る", func(t *testing.T) {
		rec := gakuseikai.do(http.MethodPost, fmt.Sprintf("/manage/stage/blocks/%d/next", bid), nil)
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		var b handler.StageBlockResponse
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &b))
		assert.Equal(t, 1, b.CurrentOrder)
		assert.True(t, b.NowPlaying)
		assert.Len(t, b.Performers, 3)
	})

	t.Run("GET /stage は認証なしで入れ子を返し、時刻は +09:00", func(t *testing.T) {
		rec := public.do(http.MethodGet, "/stage", nil)
		require.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), `"start_time":"2026-10-31T13:00:00+09:00"`)
		var res handler.StageResponse
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))
		require.Len(t, res.Sections, 1)
		s := res.Sections[0]
		assert.Equal(t, "Live1", s.Name)
		assert.Equal(t, "第一体育館", s.Location)
		require.Len(t, s.Blocks, 1)
		assert.True(t, s.Blocks[0].NowPlaying)
		assert.Equal(t, 1, s.Blocks[0].CurrentOrder)
		assert.Equal(t, "バンドB", s.Blocks[0].Performers[1].Name)
		assert.Equal(t, 2, s.Blocks[0].Performers[1].PerformOrder)
	})

	t.Run("終了から30分たてば now_playing は false(押し忘れ)", func(t *testing.T) {
		now = start.Add(50*time.Minute + domain.NowPlayingGrace)
		defer func() { now = start.Add(10 * time.Minute) }()
		var res handler.StageResponse
		require.NoError(t, json.Unmarshal(public.do(http.MethodGet, "/stage", nil).Body.Bytes(), &res))
		assert.False(t, res.Sections[0].Blocks[0].NowPlaying)
	})

	t.Run("並べ替え・前に戻す・編集", func(t *testing.T) {
		assert.Equal(t, http.StatusNoContent, admin.do(http.MethodPost, fmt.Sprintf("/manage/stage/performers/%d/move", pids[2]), map[string]string{"direction": "up"}).Code)
		assert.Equal(t, http.StatusOK, admin.do(http.MethodPost, fmt.Sprintf("/manage/stage/blocks/%d/prev", bid), nil).Code)
		assert.Equal(t, http.StatusNoContent, admin.do(http.MethodPut, fmt.Sprintf("/manage/stage/sections/%d", sid), map[string]any{"name": "Live 1", "location": "中庭", "sort_order": 1}).Code)
		assert.Equal(t, http.StatusNoContent, admin.do(http.MethodPut, fmt.Sprintf("/manage/stage/blocks/%d", bid), map[string]any{"start_time": "2026-10-31T13:00:00+09:00", "end_time": "2026-10-31T14:00:00+09:00"}).Code)
		assert.Equal(t, http.StatusNoContent, admin.do(http.MethodPut, fmt.Sprintf("/manage/stage/performers/%d", pids[0]), map[string]any{"name": "バンドA'", "detail": "", "thumbnail_url": "https://example.com/a.jpg"}).Code)

		var res handler.StageResponse
		require.NoError(t, json.Unmarshal(public.do(http.MethodGet, "/stage", nil).Body.Bytes(), &res))
		s := res.Sections[0]
		assert.Equal(t, "Live 1", s.Name)
		assert.Equal(t, "2026-10-31T14:00:00+09:00", s.Blocks[0].EndTime.Format(time.RFC3339))
		assert.Equal(t, 0, s.Blocks[0].CurrentOrder)
		var got []string
		for _, p := range s.Blocks[0].Performers {
			got = append(got, p.Name)
		}
		assert.Equal(t, []string{"バンドA'", "バンドC", "バンドB"}, got)
		assert.Equal(t, "https://example.com/a.jpg", s.Blocks[0].Performers[0].ThumbnailURL)
	})

	t.Run("権限・入力・存在のエラー", func(t *testing.T) {
		assert.Equal(t, http.StatusForbidden, student.do(http.MethodPost, fmt.Sprintf("/manage/stage/blocks/%d/next", bid), nil).Code)
		assert.Equal(t, http.StatusForbidden, gakuseikai.do(http.MethodPost, "/manage/stage/sections", map[string]any{"name": "x"}).Code)
		assert.Equal(t, http.StatusUnauthorized, public.do(http.MethodPost, fmt.Sprintf("/manage/stage/blocks/%d/next", bid), nil).Code)
		assert.Equal(t, http.StatusBadRequest, admin.do(http.MethodPost, "/manage/stage/sections", map[string]any{"name": " "}).Code)
		assert.Equal(t, http.StatusBadRequest, admin.do(http.MethodPost, fmt.Sprintf("/manage/stage/sections/%d/blocks", sid), map[string]any{"start_time": "2026-10-31T14:00:00+09:00", "end_time": "2026-10-31T13:00:00+09:00"}).Code)
		assert.Equal(t, http.StatusBadRequest, admin.do(http.MethodPost, fmt.Sprintf("/manage/stage/performers/%d/move", pids[0]), map[string]string{"direction": "left"}).Code)
		assert.Equal(t, http.StatusBadRequest, admin.do(http.MethodDelete, "/manage/stage/sections/abc", nil).Code)
		assert.Equal(t, http.StatusNotFound, gakuseikai.do(http.MethodPost, "/manage/stage/blocks/999999/next", nil).Code)
		assert.Equal(t, http.StatusNotFound, admin.do(http.MethodPut, "/manage/stage/sections/999999", map[string]any{"name": "x"}).Code)
		assert.Equal(t, http.StatusNotFound, admin.do(http.MethodPost, "/manage/stage/sections/999999/blocks", map[string]any{"start_time": "2026-10-31T13:00:00+09:00", "end_time": "2026-10-31T14:00:00+09:00"}).Code)
		assert.Equal(t, http.StatusNotFound, admin.do(http.MethodDelete, "/manage/stage/performers/999999", nil).Code)
	})

	t.Run("削除", func(t *testing.T) {
		assert.Equal(t, http.StatusNoContent, admin.do(http.MethodDelete, fmt.Sprintf("/manage/stage/performers/%d", pids[1]), nil).Code)
		assert.Equal(t, http.StatusNoContent, admin.do(http.MethodDelete, fmt.Sprintf("/manage/stage/blocks/%d", bid), nil).Code)
		assert.Equal(t, http.StatusNoContent, admin.do(http.MethodDelete, fmt.Sprintf("/manage/stage/sections/%d", sid), nil).Code)
		assert.JSONEq(t, `{"sections":[]}`, public.do(http.MethodGet, "/stage", nil).Body.String())
	})
}
