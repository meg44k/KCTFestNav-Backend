package handler_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"context"

	_ "github.com/go-sql-driver/mysql"
	"github.com/labstack/echo/v5"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"

	"github.com/meg44k/KCTFestNav-Backend/internal/domain"
	"github.com/meg44k/KCTFestNav-Backend/internal/handler"
	"github.com/meg44k/KCTFestNav-Backend/internal/repository"
	"github.com/meg44k/KCTFestNav-Backend/internal/usecase"
)

// setupUserE2ETest はテスト用のDBとEchoルーターを初期化して返します
func setupUserE2ETest(t *testing.T) (*echo.Echo, *sql.DB, *redis.Client) {
	dsn := "root:@tcp(127.0.0.1:3306)/kctfest_test?parseTime=true"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatalf("DBの初期化エラー: %v", err)
	}

	if err := db.Ping(); err != nil {
		t.Skipf("テスト用DBが起動していないためスキップします: %v", err)
	}

	// テーブルを初期化（usersテーブルを空にする）
	_, _ = db.Exec("SET FOREIGN_KEY_CHECKS = 0")
	_, err = db.Exec("TRUNCATE TABLE users")
	_, _ = db.Exec("SET FOREIGN_KEY_CHECKS = 1")
	if err != nil {
		t.Fatalf("usersテーブルの初期化に失敗しました: %v", err)
	}

	// Redisの初期化
	rdb := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
	})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		t.Skipf("テスト用Redisが起動していないためスキップします: %v", err)
	}

	// 依存関係（DI）のセットアップ
	repo := repository.NewUserRepository(db, rdb)
	uc := usecase.NewUserUsecase(repo)
	
	jwtSecret := []byte("test-secret")
	userHandler := handler.NewUserHandler(uc, jwtSecret)

	e := echo.New()
	e.HTTPErrorHandler = handler.CustomHTTPErrorHandler

	// 環境変数にJWTシークレットをセット（Loginハンドラで利用されるため）
	os.Setenv("JWT_SECRET", "test-secret")

	// ルーティングの登録
	manage := e.Group("/manage")
	manage.POST("/users", userHandler.Create)

	authGroup := e.Group("/auth")
	authGroup.POST("/login", userHandler.Login)

	return e, db, rdb
}

func TestUserE2E(t *testing.T) {
	e, db, rdb := setupUserE2ETest(t)
	defer db.Close()
	defer rdb.Close()

	t.Run("POST /manage/users - ユーザーの作成", func(t *testing.T) {
		reqBody := handler.CreateRequest{
			Name:            "E2Eテストユーザー",
			LoginID:         "e2e_test_user",
			Password:        "my_secure_password",
			AssignedBoothID: 0,
			Role:            domain.RoleStudent,
		}
		bodyBytes, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/manage/users", bytes.NewReader(bodyBytes))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		// 201 Created が返ってくること
		if !assert.Equal(t, http.StatusCreated, rec.Code) {
			t.Logf("Response body: %s", rec.Body.String())
		}

		// 実際にDBに登録されたか、パスワードがハッシュ化されているかを確認
		var dbName, dbPassword string
		err := db.QueryRow("SELECT name, password FROM users WHERE login_id = ?", "e2e_test_user").Scan(&dbName, &dbPassword)
		assert.NoError(t, err)
		assert.Equal(t, "E2Eテストユーザー", dbName)
		// 生のパスワードのまま保存されていないことを確認！
		assert.NotEqual(t, "my_secure_password", dbPassword)
	})

	t.Run("POST /auth/login - ログイン成功（正しいパスワード）", func(t *testing.T) {
		reqBody := handler.LoginRequest{
			LoginID:  "e2e_test_user",
			Password: "my_secure_password", // 先ほど登録した生のパスワード
		}
		bodyBytes, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(bodyBytes))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		// 200 OK が返ってくること
		if !assert.Equal(t, http.StatusOK, rec.Code) {
			t.Logf("Response body: %s", rec.Body.String())
		}

		var res handler.LoginResponse
		err := json.Unmarshal(rec.Body.Bytes(), &res)
		assert.NoError(t, err)

		// トークンが空ではなく、正しく返ってきていること
		assert.NotEmpty(t, res.Token)
	})

	t.Run("POST /auth/login - ログイン失敗（間違ったパスワード）", func(t *testing.T) {
		reqBody := handler.LoginRequest{
			LoginID:  "e2e_test_user",
			Password: "wrong_password!", // わざと間違える
		}
		bodyBytes, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(bodyBytes))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		// 401 Unauthorized が返ってくること
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})
	
	t.Run("POST /auth/login - ログイン失敗（存在しないLoginID）", func(t *testing.T) {
		reqBody := handler.LoginRequest{
			LoginID:  "ghost_user", // 存在しないユーザー
			Password: "my_secure_password",
		}
		bodyBytes, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(bodyBytes))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		// こちらも同じく 401 Unauthorized が返ってくること
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})
}
