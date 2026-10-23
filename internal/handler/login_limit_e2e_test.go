package handler_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/labstack/echo/v5"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/meg44k/KCTFestNav-Backend/internal/handler"
	"github.com/meg44k/KCTFestNav-Backend/internal/repository"
	"github.com/meg44k/KCTFestNav-Backend/internal/router"
	"github.com/meg44k/KCTFestNav-Backend/internal/usecase"
)

const loginKey = "e2e-login-key"

// 本物のログインのルート(合言葉・回数の上限付き)と、ログインできる管理者 1 人
func setupLoginLimitE2E(t *testing.T) *echo.Echo {
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
	require.NoError(t, truncateTables(db, "users"))

	uc := usecase.NewUserUsecase(repository.NewUserRepository(db, rdb)).WithLoginLimiter(repository.NewLoginLimiter(rdb))
	require.NoError(t, uc.InitAdminUser(context.Background(), "e2e-admin", []byte("right-password")))
	e := echo.New()
	e.HTTPErrorHandler = handler.CustomHTTPErrorHandler
	router.AuthRoutes(e, handler.NewUserHandler(uc, []byte("test-secret")), []byte("test-secret"), loginKey)
	return e
}

func login(e *echo.Echo, key, client, password string) *httptest.ResponseRecorder {
	body, _ := json.Marshal(handler.LoginRequest{LoginID: "e2e-admin", Password: password})
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	if key != "" {
		req.Header.Set("X-Internal-Key", key)
	}
	req.Header.Set("X-Login-Client", client)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

func TestLoginLimitE2E_合言葉が無いとログインを受け付けない(t *testing.T) {
	e := setupLoginLimitE2E(t)
	// Next.js のサーバーを通らない(BotID をすり抜ける)直接の呼び出しは断る
	assert.Equal(t, http.StatusUnauthorized, login(e, "", "browser-a", "right-password").Code)
	assert.Equal(t, http.StatusOK, login(e, loginKey, "browser-a", "right-password").Code)
}

func TestLoginLimitE2E_あと何回かを返し_5回で止め_別のブラウザは通す(t *testing.T) {
	e := setupLoginLimitE2E(t)
	for want := 4; want >= 1; want-- {
		rec := login(e, loginKey, "browser-a", "wrong")
		require.Equal(t, http.StatusUnauthorized, rec.Code)
		assert.JSONEq(t, `{"error":"unauthorized","remaining":`+string(rune('0'+want))+`}`, rec.Body.String())
	}
	assert.Equal(t, http.StatusTooManyRequests, login(e, loginKey, "browser-a", "wrong").Code, "5 回目で止める")
	assert.Equal(t, http.StatusTooManyRequests, login(e, loginKey, "browser-a", "right-password").Code, "止めている間は正しくても断る")
	assert.Equal(t, http.StatusOK, login(e, loginKey, "browser-b", "right-password").Code, "失敗していないブラウザの本人は通す(同じ Wi-Fi でも)")
}

func TestLoginLimitE2E_ブラウザの番号が無ければ接続元で数える(t *testing.T) {
	// 出すときの順番の間(古いフロントは接続元だけを渡す)も、全員がまとめて止まらないように
	e := setupLoginLimitE2E(t)
	old := func(ip, password string) int {
		body, _ := json.Marshal(handler.LoginRequest{LoginID: "e2e-admin", Password: password})
		req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("X-Internal-Key", loginKey)
		req.Header.Set("X-Client-IP", ip)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		return rec.Code
	}
	for i := 0; i < 5; i++ {
		old("203.0.113.1", "wrong")
	}
	assert.Equal(t, http.StatusTooManyRequests, old("203.0.113.1", "right-password"))
	assert.Equal(t, http.StatusOK, old("198.51.100.2", "right-password"), "ほかの接続元は止めない")
}

