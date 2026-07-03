package handler_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/labstack/echo/v5"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"

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
	manage.Use(mv.JWTAuth(jwtSecret))
	manage.POST("/users", userHandler.Create)
	manage.GET("/users/:id", userHandler.GetByID)
	manage.GET("/users", userHandler.GetAll)
	manage.PUT("/users/:id", userHandler.Update)
	manage.DELETE("/users/:id", userHandler.Delete)

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
	var adminToken string

	t.Run("初期化: AdminユーザーをDBに直接作成し、ログインしてトークンを取得", func(t *testing.T) {
		adminHash, _ := bcrypt.GenerateFromPassword([]byte("admin_pass"), bcrypt.DefaultCost)
		_, err := db.Exec("INSERT INTO users (id, name, login_id, password, role, assigned_booth_id) VALUES (?, ?, ?, ?, ?, NULL)", "00000000-0000-0000-0000-000000000001", "管理者", "admin_user", adminHash, domain.RoleAdmin)
		assert.NoError(t, err)

		loginReq := handler.LoginRequest{
			LoginID:  "admin_user",
			Password: "admin_pass",
		}
		loginBody, _ := json.Marshal(loginReq)
		req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(loginBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		var loginRes handler.LoginResponse
		json.Unmarshal(rec.Body.Bytes(), &loginRes)
		adminToken = loginRes.Token
	})

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
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+adminToken)
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
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+adminToken)
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

	t.Run("PUT /manage/users/:id - ユーザー情報を更新できること(Admin権限)", func(t *testing.T) {
		// 3. 更新用ブースを用意
		_, _ = db.Exec("INSERT INTO booths (id, name, organizer, detail, x, y, z) VALUES (999, 'E2Eブース', '主催', '詳細', 0, 0, 0)")

		// 4. 先ほど作成した一般ユーザー(createdUserID)の情報を更新する
		updateReq := map[string]interface{}{
			"name":              "E2Eテストユーザー(更新済)",
			"login_id":          "e2e_test_user_updated",
			"password":          "new_secure_password", // 生の文字列
			"assigned_booth_id": 999,
			"role":              domain.RoleMember,
		}
		updateBody, _ := json.Marshal(updateReq)
		req3 := httptest.NewRequest(http.MethodPut, "/manage/users/"+createdUserID, bytes.NewReader(updateBody))
		req3.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req3.Header.Set(echo.HeaderAuthorization, "Bearer "+adminToken)
		rec3 := httptest.NewRecorder()
		e.ServeHTTP(rec3, req3)

		if !assert.Equal(t, http.StatusNoContent, rec3.Code) {
			t.Logf("Response body: %s", rec3.Body.String())
		}

		// 5. 更新されたことを確認
		var dbName string
		var dbRole string
		var dbBoothID sql.NullInt32
		err := db.QueryRow("SELECT name, role, assigned_booth_id FROM users WHERE id = ?", createdUserID).Scan(&dbName, &dbRole, &dbBoothID)
		assert.NoError(t, err)
		assert.Equal(t, "E2Eテストユーザー(更新済)", dbName)
		assert.Equal(t, string(domain.RoleMember), dbRole)
		assert.True(t, dbBoothID.Valid)
		assert.Equal(t, int32(999), dbBoothID.Int32)
	})

	t.Run("DELETE /manage/users/:id - ユーザーを削除できること(Admin権限)", func(t *testing.T) {
		// Adminユーザーでログインしてトークンを取得
		loginReq := handler.LoginRequest{
			LoginID:  "admin_user",
			Password: "admin_pass",
		}
		loginBody, _ := json.Marshal(loginReq)
		reqLogin := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(loginBody))
		reqLogin.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		recLogin := httptest.NewRecorder()
		e.ServeHTTP(recLogin, reqLogin)
		assert.Equal(t, http.StatusOK, recLogin.Code)

		var loginRes handler.LoginResponse
		json.Unmarshal(recLogin.Body.Bytes(), &loginRes)
		adminToken := loginRes.Token

		// Deleteリクエスト
		reqDel := httptest.NewRequest(http.MethodDelete, "/manage/users/"+createdUserID, nil)
		reqDel.Header.Set(echo.HeaderAuthorization, "Bearer "+adminToken)
		recDel := httptest.NewRecorder()
		e.ServeHTTP(recDel, reqDel)

		if !assert.Equal(t, http.StatusNoContent, recDel.Code) {
			t.Logf("Response body: %s", recDel.Body.String())
		}

		// 削除されたかDBで確認
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM users WHERE id = ?", createdUserID).Scan(&count)
		assert.NoError(t, err)
		assert.Equal(t, 0, count)
	})
}
