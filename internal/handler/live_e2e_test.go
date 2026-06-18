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
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"

	"github.com/meg44k/KCTFestNav-Backend/internal/domain"
	"github.com/meg44k/KCTFestNav-Backend/internal/handler"
	"github.com/meg44k/KCTFestNav-Backend/internal/repository"
	"github.com/meg44k/KCTFestNav-Backend/internal/router"
	"github.com/meg44k/KCTFestNav-Backend/internal/usecase"
	"github.com/redis/go-redis/v9"
)

// setupE2ETest はテスト用のDBとEchoルーターを初期化して返します
func setupE2ETest(t *testing.T) (*echo.Echo, *sql.DB, *redis.Client) {
	dsn := "root:@tcp(127.0.0.1:3306)/kctfest_test?parseTime=true"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatalf("DBの初期化エラー: %v", err)
	}

	if err := db.Ping(); err != nil {
		t.Skipf("テスト用DBが起動していないためスキップします: %v", err)
	}

	// テーブルを初期化
	_, _ = db.Exec("TRUNCATE TABLE lives")

	// Redisの初期化
	rdb := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
	})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		t.Skipf("テスト用Redisが起動していないためスキップします: %v", err)
	}
	rdb.Del(context.Background(), "lives:current")

	// 依存関係（DI）のセットアップ
	repo := repository.NewLiveRepository(db, rdb)
	uc := usecase.NewLiveUsecase(repo)
	liveHandler := handler.NewLiveHandler(uc)

	e := echo.New()

	// Usecase側で必要になる「管理者権限（RoleAdmin）」をコンテキストに強制注入するモックミドルウェア
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			// context.ContextにUserRoleKeyをセットする
			ctx := context.WithValue(c.Request().Context(), domain.ContextUserRoleKey, domain.RoleAdmin)
			c.SetRequest(c.Request().WithContext(ctx))
			return next(c)
		}
	})

	// カスタムエラーハンドラを設定
	e.HTTPErrorHandler = handler.CustomHTTPErrorHandler

	// ルーティングの登録
	router.InitRoutes(e, &handler.Handlers{Live: liveHandler})

	return e, db, rdb
}

func TestLiveE2E(t *testing.T) {
	e, db, rdb := setupE2ETest(t)
	defer db.Close()
	defer rdb.Close()

	// テスト間でデータを共有するためにIDを保持
	var insertedLiveID int

	t.Run("POST /manage/lives - ライブの作成", func(t *testing.T) {
		reqBody := handler.CreateLiveRequest{
			Name:          "E2Eテストライブ",
			Detail:        "E2Eの詳細",
			ThumbnailURL:  "https://example.com/e2e.jpg",
			StartTime:     time.Now().Add(1 * time.Hour).Round(time.Second),
			EndTime:       time.Now().Add(2 * time.Hour).Round(time.Second),
			SessionNumber: 1,
		}
		bodyBytes, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/manage/lives", bytes.NewReader(bodyBytes))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		// 201 Created が返ってくること
		assert.Equal(t, http.StatusCreated, rec.Code)

		// 実際にDBに登録されたIDを取得して次に回す
		err := db.QueryRow("SELECT id FROM lives LIMIT 1").Scan(&insertedLiveID)
		assert.NoError(t, err)
	})

	t.Run("GET /lives - 全件取得", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/lives", nil)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var res handler.GetAllLivesResponse
		err := json.Unmarshal(rec.Body.Bytes(), &res)
		assert.NoError(t, err)

		// さっき作成した1件が入っているはず
		assert.Len(t, res.Lives, 1)
		assert.Equal(t, "E2Eテストライブ", res.Lives[0].Name)
	})

	t.Run("GET /lives/:id - 1件取得", func(t *testing.T) {
		// さっき作られた実際のIDを使ってリクエスト
		path := fmt.Sprintf("/lives/%d", insertedLiveID)
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var res handler.LiveResponse
		err := json.Unmarshal(rec.Body.Bytes(), &res)
		assert.NoError(t, err)

		assert.Equal(t, insertedLiveID, res.ID)
		assert.Equal(t, domain.LiveStatus(0), res.Status)
		assert.Equal(t, "E2Eテストライブ", res.Name)
	})

	t.Run("PUT /manage/lives/:id - ライブの更新", func(t *testing.T) {
		reqBody := handler.UpdateLiveRequest{
			Name:          "E2E更新済みライブ",
			Detail:        "E2E更新済み詳細",
			ThumbnailURL:  "https://example.com/updated.jpg",
			StartTime:     time.Now().Add(1 * time.Hour).Round(time.Second),
			EndTime:       time.Now().Add(2 * time.Hour).Round(time.Second),
			SessionNumber: 2,
			Status:        1,
		}
		bodyBytes, _ := json.Marshal(reqBody)

		path := fmt.Sprintf("/manage/lives/%d", insertedLiveID)
		req := httptest.NewRequest(http.MethodPut, path, bytes.NewReader(bodyBytes))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		// 204 NoContent が返ってくること
		assert.Equal(t, http.StatusNoContent, rec.Code)

		// 本当に更新されたかDBを見て確認
		var newName string
		err := db.QueryRow("SELECT name FROM lives WHERE id = ?", insertedLiveID).Scan(&newName)
		assert.NoError(t, err)
		assert.Equal(t, "E2E更新済みライブ", newName)
	})

	t.Run("GET /lives/current - 現在進行中のライブを取得", func(t *testing.T) {
		// 直前のPUTでStatusを1（進行中）に更新したので、ここで取得できるはず
		req := httptest.NewRequest(http.MethodGet, "/lives/current", nil)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		// 200 OK が返ってくること
		assert.Equal(t, http.StatusOK, rec.Code)

		var res handler.LiveResponse
		err := json.Unmarshal(rec.Body.Bytes(), &res)
		assert.NoError(t, err)

		// さっき更新した「E2E更新済みライブ」が進行中として取得できること
		assert.Equal(t, insertedLiveID, res.ID)
		assert.Equal(t, "E2E更新済みライブ", res.Name)
		assert.Equal(t, domain.LiveStatus(1), res.Status)
	})

	t.Run("DELETE /manage/lives/:id - ライブの削除", func(t *testing.T) {
		path := fmt.Sprintf("/manage/lives/%d", insertedLiveID)
		req := httptest.NewRequest(http.MethodDelete, path, nil)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		// 204 NoContent が返ってくること
		assert.Equal(t, http.StatusNoContent, rec.Code)

		// 本当に削除されたか確認
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM lives WHERE id = ?", insertedLiveID).Scan(&count)
		assert.NoError(t, err)
		assert.Equal(t, 0, count)
	})

	t.Run("GET /lives/current - 進行中がない場合の確認", func(t *testing.T) {
		// 直前のDELETEでデータが消えたので、進行中のライブは見つからないはず
		req := httptest.NewRequest(http.MethodGet, "/lives/current", nil)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		// 404 NotFound が返ってくること
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})
}
