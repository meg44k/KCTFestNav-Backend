package handler_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/labstack/echo/v5"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/meg44k/KCTFestNav-Backend/internal/domain"
	"github.com/meg44k/KCTFestNav-Backend/internal/handler"
	"github.com/meg44k/KCTFestNav-Backend/internal/middleware"
	"github.com/meg44k/KCTFestNav-Backend/internal/repository"
	"github.com/meg44k/KCTFestNav-Backend/internal/router"
	"github.com/meg44k/KCTFestNav-Backend/internal/usecase"
)

const likeKey = "e2e-internal-key"

type likeEnv struct {
	t      *testing.T
	e      *echo.Echo
	secret []byte
	class  int
	bazaar int
}

// 本物のルーターに、いいねの handler をつなぐ。クラス展示とバザーを 1 つずつ作る
func setupLikeE2E(t *testing.T) likeEnv {
	t.Helper()
	db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/kctfest_test_handler?parseTime=true")
	require.NoError(t, err)
	if err := db.Ping(); err != nil {
		t.Skipf("テスト用DBが起動していないためスキップします: %v", err)
	}
	rdb := redis.NewClient(&redis.Options{Addr: "127.0.0.1:6379", DB: 1})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		t.Skipf("テスト用Redisが起動していないためスキップします: %v", err)
	}
	t.Cleanup(func() { db.Close(); rdb.Close() })
	require.NoError(t, rdb.FlushDB(context.Background()).Err())
	for _, q := range []string{"SET FOREIGN_KEY_CHECKS = 0", "TRUNCATE TABLE likes", "TRUNCATE TABLE like_removals", "TRUNCATE TABLE booths", "SET FOREIGN_KEY_CHECKS = 1"} {
		_, err := db.Exec(q)
		require.NoError(t, err)
	}
	insert := func(name, org string) int {
		res, err := db.Exec(`INSERT INTO booths (name, organizer, detail, x, y, z) VALUES (?, ?, '', 0, 0, 0)`, name, org)
		require.NoError(t, err)
		id, _ := res.LastInsertId()
		return int(id)
	}
	env := likeEnv{t: t, secret: []byte("e2e-test-secret")}
	env.class = insert("お化け屋敷", "3-2")
	env.bazaar = insert("からあげ", "天文部")

	uc := usecase.NewLikeUsecase(repository.NewLikeRepository(db, rdb), repository.NewBoothRepository(db, rdb), []byte("voter-secret"), time.Now)
	e := echo.New()
	e.HTTPErrorHandler = handler.CustomHTTPErrorHandler
	router.LikeRoutes(e, e.Group("/manage", middleware.JWTAuth(env.secret)), handler.NewLikeHandler(uc), likeKey)
	env.e = e
	return env
}

func (env likeEnv) call(method, path string, headers map[string]string) *httptest.ResponseRecorder {
	env.t.Helper()
	req := httptest.NewRequest(method, path, nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	env.e.ServeHTTP(rec, req)
	return rec
}

func (env likeEnv) voter() string {
	env.t.Helper()
	rec := env.call(http.MethodPost, "/internal/voters", map[string]string{"X-Internal-Key": likeKey})
	require.Equal(env.t, http.StatusCreated, rec.Code, rec.Body.String())
	var res struct {
		Voter string `json:"voter"`
	}
	require.NoError(env.t, json.Unmarshal(rec.Body.Bytes(), &res))
	return res.Voter
}

func voterHeaders(v string) map[string]string {
	return map[string]string{"X-Internal-Key": likeKey, "X-Voter": v}
}

func (env likeEnv) bearer(role domain.Role) map[string]string {
	return map[string]string{"Authorization": "Bearer " + tokenFor(env.t, role, env.secret)}
}

type likesSummary struct {
	Booths []struct {
		BoothID int  `json:"booth_id"`
		Total   int  `json:"total"`
		Burst   bool `json:"burst"`
		Buckets []struct {
			Start string `json:"start"`
			Count int    `json:"count"`
		} `json:"buckets"`
	} `json:"booths"`
}

func (env likeEnv) summary() likesSummary {
	env.t.Helper()
	rec := env.call(http.MethodGet, "/manage/likes", env.bearer(domain.RoleGakuseikai))
	require.Equal(env.t, http.StatusOK, rec.Code, rec.Body.String())
	var s likesSummary
	require.NoError(env.t, json.Unmarshal(rec.Body.Bytes(), &s))
	return s
}

func TestLikeE2E_合言葉が無いと呼べない(t *testing.T) {
	env := setupLikeE2E(t)
	assert.Equal(t, http.StatusUnauthorized, env.call(http.MethodPost, "/internal/voters", nil).Code)
	assert.Equal(t, http.StatusUnauthorized, env.call(http.MethodPost, "/internal/voters", map[string]string{"X-Internal-Key": "x"}).Code)
	v := env.voter()
	assert.Equal(t, http.StatusUnauthorized, env.call(http.MethodPut, fmt.Sprintf("/internal/likes/%d", env.class), map[string]string{"X-Voter": v}).Code)
}

func TestLikeE2E_押して自分のいいねに入り_2回でも1件(t *testing.T) {
	env := setupLikeE2E(t)
	v := env.voter()
	path := fmt.Sprintf("/internal/likes/%d", env.class)
	assert.Equal(t, http.StatusNoContent, env.call(http.MethodPut, path, voterHeaders(v)).Code)
	assert.Equal(t, http.StatusNoContent, env.call(http.MethodPut, path, voterHeaders(v)).Code)

	rec := env.call(http.MethodGet, "/internal/likes/mine", voterHeaders(v))
	require.Equal(t, http.StatusOK, rec.Code)
	assert.JSONEq(t, fmt.Sprintf(`{"booth_ids":[%d]}`, env.class), rec.Body.String())

	s := env.summary()
	require.Len(t, s.Booths, 1, "クラス展示だけ")
	assert.Equal(t, 1, s.Booths[0].Total)

	assert.Equal(t, http.StatusNoContent, env.call(http.MethodDelete, path, voterHeaders(v)).Code)
	rec = env.call(http.MethodGet, "/internal/likes/mine", voterHeaders(v))
	assert.JSONEq(t, `{"booth_ids":[]}`, rec.Body.String())
}

func TestLikeE2E_番号が無いときの自分のいいねは空(t *testing.T) {
	env := setupLikeE2E(t)
	rec := env.call(http.MethodGet, "/internal/likes/mine", map[string]string{"X-Internal-Key": likeKey})
	require.Equal(t, http.StatusOK, rec.Code)
	assert.JSONEq(t, `{"booth_ids":[]}`, rec.Body.String())
}

func TestLikeE2E_押せないもの(t *testing.T) {
	env := setupLikeE2E(t)
	v := env.voter()
	assert.Equal(t, http.StatusBadRequest, env.call(http.MethodPut, fmt.Sprintf("/internal/likes/%d", env.bazaar), voterHeaders(v)).Code, "バザー")
	assert.Equal(t, http.StatusNotFound, env.call(http.MethodPut, "/internal/likes/99999", voterHeaders(v)).Code, "無いブース")
	assert.Equal(t, http.StatusUnauthorized, env.call(http.MethodPut, fmt.Sprintf("/internal/likes/%d", env.class), voterHeaders(v+"x")).Code, "書き換えた番号")
	assert.Equal(t, http.StatusUnauthorized, env.call(http.MethodPut, fmt.Sprintf("/internal/likes/%d", env.class), map[string]string{"X-Internal-Key": likeKey}).Code, "番号なし")
}

func TestLikeE2E_管理用の権限(t *testing.T) {
	env := setupLikeE2E(t)
	assert.Equal(t, http.StatusUnauthorized, env.call(http.MethodGet, "/manage/likes", nil).Code)
	assert.Equal(t, http.StatusForbidden, env.call(http.MethodGet, "/manage/likes", env.bearer(domain.RoleStudent)).Code)
	assert.Equal(t, http.StatusForbidden, env.call(http.MethodDelete, "/manage/likes", env.bearer(domain.RoleGakuseikai)).Code, "全部消すは管理者だけ")
	rec := env.call(http.MethodDelete, "/manage/likes", env.bearer(domain.RoleAdmin))
	require.Equal(t, http.StatusOK, rec.Code)
	assert.JSONEq(t, `{"removed":0}`, rec.Body.String())
}

func TestLikeE2E_時間帯で取り消す(t *testing.T) {
	env := setupLikeE2E(t)
	v := env.voter()
	require.Equal(t, http.StatusNoContent, env.call(http.MethodPut, fmt.Sprintf("/internal/likes/%d", env.class), voterHeaders(v)).Code)

	now := time.Now().UTC()
	q := url.Values{"from": {now.Add(-time.Hour).Format(time.RFC3339)}, "to": {now.Add(time.Hour).Format(time.RFC3339)}}
	rec := env.call(http.MethodDelete, fmt.Sprintf("/manage/likes/%d?%s", env.class, q.Encode()), env.bearer(domain.RoleGakuseikai))
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	assert.JSONEq(t, `{"removed":1}`, rec.Body.String())
	assert.Equal(t, 0, env.summary().Booths[0].Total)

	bad := url.Values{"from": {now.Format(time.RFC3339)}, "to": {now.Format(time.RFC3339)}}
	assert.Equal(t, http.StatusBadRequest, env.call(http.MethodDelete, fmt.Sprintf("/manage/likes/%d?%s", env.class, bad.Encode()), env.bearer(domain.RoleGakuseikai)).Code, "from >= to")
	assert.Equal(t, http.StatusBadRequest, env.call(http.MethodDelete, fmt.Sprintf("/manage/likes/%d?from=x&to=y", env.class), env.bearer(domain.RoleGakuseikai)).Code, "時刻の形が違う")
}
