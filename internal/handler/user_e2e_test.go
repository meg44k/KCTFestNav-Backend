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
	mv "github.com/meg44k/KCTFestNav-Backend/internal/middleware"
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
	manage.GET("/users/:id", userHandler.GetByID)
	manage.GET("/users", userHandler.GetAll)

	authGroup := e.Group("/auth")
	authGroup.POST("/login", userHandler.Login)
	authGroup.GET("/me", userHandler.GetMe, mv.JWTAuth(jwtSecret))
	
	// パス確認用（パブリックな取得想定）
	e.GET("/users/:id", userHandler.GetByID)

	return e, db, rdb
}

func TestUserE2E(t *testing.T) {
	e, db, rdb := setupUserE2ETest(t)
	defer db.Close()
	defer rdb.Close()

	var loginToken string
	var createdUserID string

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
		var dbID string
		var dbName, dbPassword string
		err := db.QueryRow("SELECT id, name, password FROM users WHERE login_id = ?", "e2e_test_user").Scan(&dbID, &dbName, &dbPassword)
		assert.NoError(t, err)
		assert.Equal(t, "E2Eテストユーザー", dbName)
		
		createdUserID = dbID
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
		loginToken = res.Token
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

	t.Run("GET /auth/me - 自身の情報を取得できること", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+loginToken)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		if !assert.Equal(t, http.StatusOK, rec.Code) {
			t.Logf("Response body: %s", rec.Body.String())
		}

		var res handler.GetUserResponse
		err := json.Unmarshal(rec.Body.Bytes(), &res)
		assert.NoError(t, err)

		assert.Equal(t, "E2Eテストユーザー", res.Name)
		assert.Equal(t, "e2e_test_user", res.LoginID)
		assert.Equal(t, domain.RoleStudent, res.Role)
	})

	t.Run("GET /auth/me - トークンがない場合はエラーになること", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})
	t.Run("GET /users/:id - ユーザー情報をIDで取得できること", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/users/"+createdUserID, nil)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		if !assert.Equal(t, http.StatusOK, rec.Code) {
			t.Logf("Response body: %s", rec.Body.String())
		}

		var res handler.GetUserResponse
		err := json.Unmarshal(rec.Body.Bytes(), &res)
		assert.NoError(t, err)

		assert.Equal(t, createdUserID, res.ID.String())
		assert.Equal(t, "E2Eテストユーザー", res.Name)
		assert.Equal(t, "e2e_test_user", res.LoginID)
		assert.Equal(t, domain.RoleStudent, res.Role)
	})

	t.Run("GET /users/:id - 存在しないUUIDの場合は404エラーになること", func(t *testing.T) {
		// ランダムなUUIDを生成
		randomID := "123e4567-e89b-12d3-a456-426614174000"
		req := httptest.NewRequest(http.MethodGet, "/users/"+randomID, nil)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("GET /manage/users - ユーザー一覧を取得できること", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/manage/users", nil)
		// manage グループは JWT 認証が必要（※E2EのsetupUserE2ETestではmanageにJWTをかけていないが、一応付けておく）
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+loginToken)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		if !assert.Equal(t, http.StatusOK, rec.Code) {
			t.Logf("Response body: %s", rec.Body.String())
		}

		var res handler.GetAllUsersResponse
		err := json.Unmarshal(rec.Body.Bytes(), &res)
		assert.NoError(t, err)

		// さきほど作成したE2Eテストユーザーが最低1人は含まれているはず
		assert.GreaterOrEqual(t, len(res.Users), 1)
		
		var found bool
		for _, u := range res.Users {
			if u.LoginID == "e2e_test_user" {
				assert.Equal(t, "E2Eテストユーザー", u.Name)
				assert.Equal(t, domain.RoleStudent, u.Role)
				found = true
				break
			}
		}
		assert.True(t, found, "作成したテストユーザーが一覧に含まれていません")
	})
}
