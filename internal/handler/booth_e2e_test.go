package handler_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	_ "github.com/go-sql-driver/mysql"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/meg44k/KCTFestNav-Backend/internal/auth"
	"github.com/meg44k/KCTFestNav-Backend/internal/domain"
	"github.com/meg44k/KCTFestNav-Backend/internal/handler"
	"github.com/meg44k/KCTFestNav-Backend/internal/middleware"
	"github.com/meg44k/KCTFestNav-Backend/internal/repository"
	"github.com/meg44k/KCTFestNav-Backend/internal/usecase"
	"github.com/redis/go-redis/v9"
)

// setupBoothE2ETest はテスト用のDBとEchoルーターを初期化して返します
func setupBoothE2ETest(t *testing.T) (*echo.Echo, *sql.DB, *redis.Client, []byte) {
	dsn := "root:@tcp(127.0.0.1:3306)/kctfest_test?parseTime=true"
	db, err := sql.Open("mysql", dsn)
	require.NoError(t, err, "DBの初期化エラー")

	if err := db.Ping(); err != nil {
		t.Skipf("テスト用DBが起動していないためスキップします: %v", err)
	}

	// テーブルを初期化 (外部キー制約を一時的に無視する)
	_, _ = db.Exec("SET FOREIGN_KEY_CHECKS = 0")
	_, err = db.Exec("TRUNCATE TABLE booths")
	_, _ = db.Exec("SET FOREIGN_KEY_CHECKS = 1")
	require.NoError(t, err)

	// Redisの初期化
	rdb := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
	})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		t.Skipf("テスト用Redisが起動していないためスキップします: %v", err)
	}
	rdb.FlushDB(context.Background())

	// 依存関係（DI）のセットアップ
	repo := repository.NewBoothRepository(db, rdb)
	uc := usecase.NewBoothUsecase(repo)
	boothHandler := handler.NewBoothHandler(uc)

	e := echo.New()

	e.HTTPErrorHandler = handler.CustomHTTPErrorHandler

	jwtSecret := []byte("booth-e2e-secret")

	// ルーティングの手動登録
	manage := e.Group("/manage")
	manage.Use(middleware.JWTAuth(jwtSecret))

	manage.POST("/booths", boothHandler.Create)
	manage.PUT("/booths/:id", boothHandler.Update)
	manage.PATCH("/booths/:id/congestion", boothHandler.UpdateCongestion)
	manage.DELETE("/booths/:id", boothHandler.Delete)

	e.GET("/booths", boothHandler.GetAll)
	e.GET("/booths/:id", boothHandler.GetByID)

	return e, db, rdb, jwtSecret
}

func TestBoothE2E(t *testing.T) {
	e, db, rdb, jwtSecret := setupBoothE2ETest(t)
	defer db.Close()
	defer rdb.Close()

	// Adminのトークン
	adminTokenStr, _ := auth.GenerateToken(uuid.New(), domain.RoleAdmin, 0, jwtSecret)
	adminAuthHeader := "Bearer " + adminTokenStr

	var insertedBoothID int

	t.Run("POST /manage/booths - ブースの作成", func(t *testing.T) {
		reqBody := handler.CreateBoothRequest{
			Name:      "E2Eテストブース",
			Organizer: "E2Eテスト実行委員会",
			Detail:    "E2Eテスト用の詳細情報",
			X:         1.5,
			Y:         2.5,
			Z:         3.5,
		}
		bodyBytes, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/manage/booths", bytes.NewReader(bodyBytes))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Authorization", adminAuthHeader)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusCreated, rec.Code)

		err := db.QueryRow("SELECT id FROM booths LIMIT 1").Scan(&insertedBoothID)
		assert.NoError(t, err)
		assert.NotZero(t, insertedBoothID)
	})

	t.Run("異常系: POST /manage/booths - 名前が空文字の場合は作成できない(400)", func(t *testing.T) {
		reqBody := handler.CreateBoothRequest{
			Name:      "", // 空文字
			Organizer: "E2Eテスト実行委員会",
		}
		bodyBytes, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/manage/booths", bytes.NewReader(bodyBytes))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Authorization", adminAuthHeader)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		// 400 Bad Request が返ってくること
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("GET /booths - 全件取得", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/booths", nil)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var res handler.GetAllBoothsResponse
		err := json.Unmarshal(rec.Body.Bytes(), &res)
		assert.NoError(t, err)

		assert.Len(t, res.Booths, 1)
		assert.Equal(t, "E2Eテストブース", res.Booths[0].Name)
		assert.Equal(t, domain.CongestionStatus(0), res.Booths[0].CongestionStatus)
	})

	t.Run("GET /booths/:id - 1件取得", func(t *testing.T) {
		path := fmt.Sprintf("/booths/%d", insertedBoothID)
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var res handler.GetBoothResponse
		err := json.Unmarshal(rec.Body.Bytes(), &res)
		assert.NoError(t, err)

		assert.Equal(t, insertedBoothID, res.ID)
		assert.Equal(t, "E2Eテスト実行委員会", res.Organizer)
	})

	t.Run("PUT /manage/booths/:id - ブースの更新", func(t *testing.T) {
		reqBody := handler.UpdateBoothRequest{
			Name:      "更新済みブース",
			Organizer: "更新済みオーガナイザー",
			Detail:    "詳細も更新",
			X:         10.0,
			Y:         20.0,
			Z:         30.0,
		}
		bodyBytes, _ := json.Marshal(reqBody)

		path := fmt.Sprintf("/manage/booths/%d", insertedBoothID)
		req := httptest.NewRequest(http.MethodPut, path, bytes.NewReader(bodyBytes))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Authorization", adminAuthHeader)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var newName string
		err := db.QueryRow("SELECT name FROM booths WHERE id = ?", insertedBoothID).Scan(&newName)
		assert.NoError(t, err)
		assert.Equal(t, "更新済みブース", newName)
	})

	t.Run("PATCH /manage/booths/:id/congestion - 混雑度の変更(Admin)", func(t *testing.T) {
		// 混雑度を 2 (とても混雑している) に変更
		reqBody := handler.UpdateCongestionStatusRequest{
			CongestionStatus: 2,
		}
		bodyBytes, _ := json.Marshal(reqBody)

		path := fmt.Sprintf("/manage/booths/%d/congestion", insertedBoothID)
		req := httptest.NewRequest(http.MethodPatch, path, bytes.NewReader(bodyBytes))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Authorization", adminAuthHeader)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		// Redis に正しく登録されているか確認
		val, err := rdb.Get(context.Background(), fmt.Sprintf("congestion_status:%d", insertedBoothID)).Int()
		assert.NoError(t, err)
		assert.Equal(t, 2, val)
	})

	t.Run("PATCH /manage/booths/:id/congestion - アサインされたStudentは変更できる(200)", func(t *testing.T) {
		reqBody := handler.UpdateCongestionStatusRequest{
			CongestionStatus: 0, // 空きに戻す
		}
		bodyBytes, _ := json.Marshal(reqBody)

		path := fmt.Sprintf("/manage/booths/%d/congestion", insertedBoothID)
		req := httptest.NewRequest(http.MethodPatch, path, bytes.NewReader(bodyBytes))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		
		// 権限をStudentにし、アサイン先をこのブースIDに一致させるJWTトークンを発行！
		studentToken, _ := auth.GenerateToken(uuid.New(), domain.RoleStudent, insertedBoothID, jwtSecret)
		req.Header.Set("Authorization", "Bearer "+studentToken)
		
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		// Redis に正しく登録されているか確認
		val, err := rdb.Get(context.Background(), fmt.Sprintf("congestion_status:%d", insertedBoothID)).Int()
		assert.NoError(t, err)
		assert.Equal(t, 0, val)
	})

	t.Run("異常系: PATCH /manage/booths/:id/congestion - 担当外のStudentは変更できない(403)", func(t *testing.T) {
		reqBody := handler.UpdateCongestionStatusRequest{
			CongestionStatus: 1,
		}
		bodyBytes, _ := json.Marshal(reqBody)

		path := fmt.Sprintf("/manage/booths/%d/congestion", insertedBoothID)
		req := httptest.NewRequest(http.MethodPatch, path, bytes.NewReader(bodyBytes))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		
		// 権限をStudentにし、アサイン先を別のブースID(9999)にするJWTトークン
		wrongStudentToken, _ := auth.GenerateToken(uuid.New(), domain.RoleStudent, 9999, jwtSecret)
		req.Header.Set("Authorization", "Bearer "+wrongStudentToken)
		
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		// 403 Forbidden が返ってくること
		assert.Equal(t, http.StatusForbidden, rec.Code)
	})

	t.Run("GET /booths/:id - 更新後の混雑度が取れること", func(t *testing.T) {
		path := fmt.Sprintf("/booths/%d", insertedBoothID)
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var res handler.GetBoothResponse
		err := json.Unmarshal(rec.Body.Bytes(), &res)
		assert.NoError(t, err)

		// 混雑度が2に更新されている
		assert.Equal(t, domain.CongestionStatus(0), res.CongestionStatus)
	})

	t.Run("異常系: DELETE /manage/booths/:id - 一般学生は削除できない(403)", func(t *testing.T) {
		path := fmt.Sprintf("/manage/booths/%d", insertedBoothID)
		req := httptest.NewRequest(http.MethodDelete, path, nil)
		
		// 権限をStudentにしたJWTトークン
		studentToken, _ := auth.GenerateToken(uuid.New(), domain.RoleStudent, insertedBoothID, jwtSecret)
		req.Header.Set("Authorization", "Bearer "+studentToken)
		
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		// 403 Forbidden が返ってくること
		assert.Equal(t, http.StatusForbidden, rec.Code)
	})

	t.Run("DELETE /manage/booths/:id - ブースの削除(Admin)", func(t *testing.T) {
		path := fmt.Sprintf("/manage/booths/%d", insertedBoothID)
		req := httptest.NewRequest(http.MethodDelete, path, nil)
		req.Header.Set("Authorization", adminAuthHeader) // 明示的にAdmin
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNoContent, rec.Code)

		// DBから消えていること
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM booths WHERE id = ?", insertedBoothID).Scan(&count)
		assert.NoError(t, err)
		assert.Equal(t, 0, count)

		// 🚨 Redisからも消えていること (メモリリーク防止)
		err = rdb.Get(context.Background(), fmt.Sprintf("congestion_status:%d", insertedBoothID)).Err()
		assert.ErrorIs(t, err, redis.Nil)
	})

	t.Run("異常系: GET /booths/:id - 削除されたブースの取得(404)", func(t *testing.T) {
		path := fmt.Sprintf("/booths/%d", insertedBoothID)
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		// 404 NotFound が返ってくること
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})
}
