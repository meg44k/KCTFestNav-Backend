package handler_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/labstack/echo/v5"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/google/uuid"
	"github.com/meg44k/KCTFestNav-Backend/internal/auth"
	"github.com/meg44k/KCTFestNav-Backend/internal/domain"
	"github.com/meg44k/KCTFestNav-Backend/internal/handler"
	"github.com/meg44k/KCTFestNav-Backend/internal/middleware"
	"github.com/meg44k/KCTFestNav-Backend/internal/repository"
	"github.com/meg44k/KCTFestNav-Backend/internal/router"
	"github.com/meg44k/KCTFestNav-Backend/internal/storage"
	"github.com/meg44k/KCTFestNav-Backend/internal/usecase"
)

const imageBase = "https://example.test/images"

func setupImageE2E(t *testing.T) (*echo.Echo, []byte, string, int) {
	t.Helper()
	db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/kctfest_test_handler?parseTime=true")
	require.NoError(t, err)
	if err := db.Ping(); err != nil {
		t.Skipf("テスト用DBが起動していないためスキップします: %v", err)
	}
	rdb := redis.NewClient(&redis.Options{Addr: "127.0.0.1:6379", DB: 1})
	t.Cleanup(func() { db.Close(); rdb.Close() })
	require.NoError(t, truncateTables(db, "likes", "booths"))
	res, err := db.Exec(`INSERT INTO booths (name, organizer, detail, x, y, z) VALUES ('たこ焼き', '1-1', '', 0, 0, 0)`)
	require.NoError(t, err)
	id, _ := res.LastInsertId()

	dir := t.TempDir()
	store := storage.NewLocal(dir, imageBase)
	uc := usecase.NewImageUsecase(store, repository.NewBoothRepository(db, rdb), repository.NewStageRepository(db))
	secret := []byte("e2e-test-secret")
	e := echo.New()
	e.HTTPErrorHandler = handler.CustomHTTPErrorHandler
	router.ImageRoutes(e.Group("/manage", middleware.JWTAuth(secret)), handler.NewImageHandler(uc))
	_ = context.Background()
	return e, secret, dir, int(id)
}

func upload(t *testing.T, e *echo.Echo, token, target string, data []byte) *httptest.ResponseRecorder {
	t.Helper()
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	if target != "" {
		require.NoError(t, w.WriteField("target", target))
	}
	if data != nil {
		f, err := w.CreateFormFile("file", "photo.png")
		require.NoError(t, err)
		_, _ = f.Write(data)
	}
	require.NoError(t, w.Close())
	req := httptest.NewRequest(http.MethodPost, "/manage/images", &body)
	req.Header.Set(echo.HeaderContentType, w.FormDataContentType())
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+token)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

func TestImageE2E(t *testing.T) {
	e, secret, dir, boothID := setupImageE2E(t)
	png := []byte("\x89PNG\r\n\x1a\n" + strings.Repeat("\x00", 32))
	admin, _ := auth.GenerateToken(uuid.New(), domain.RoleAdmin, 0, secret)
	other, _ := auth.GenerateToken(uuid.New(), domain.RoleStudent, boothID+1, secret)

	rec := upload(t, e, admin, fmt.Sprintf("booth:%d", boothID), png)
	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
	var res struct {
		URL string `json:"url"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))
	prefix := fmt.Sprintf("%s/booths/%d/", imageBase, boothID)
	require.True(t, strings.HasPrefix(res.URL, prefix), res.URL)
	saved, err := os.ReadFile(filepath.Join(dir, strings.TrimPrefix(res.URL, imageBase+"/")))
	require.NoError(t, err)
	assert.Equal(t, png, saved)

	assert.Equal(t, http.StatusForbidden, upload(t, e, other, fmt.Sprintf("booth:%d", boothID), png).Code, "ほかのブースの担当")
	assert.Equal(t, http.StatusBadRequest, upload(t, e, admin, "", png).Code, "target なし")
	assert.Equal(t, http.StatusBadRequest, upload(t, e, admin, fmt.Sprintf("booth:%d", boothID), []byte("<html>")).Code, "画像でない")
	assert.Equal(t, http.StatusBadRequest, upload(t, e, admin, fmt.Sprintf("booth:%d", boothID), nil).Code, "file なし")
	assert.Equal(t, http.StatusNotFound, upload(t, e, admin, "booth:99999", png).Code, "無いブース")
}
