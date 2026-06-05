package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/meg44k/KCTFestNav-Backend/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockLiveUsecase struct {
	createFn  func(ctx context.Context, name string, detail string, thumbnailURL string, startTime time.Time, endTime time.Time, sessionNumber int8) error
	updateFn  func(ctx context.Context, id int, name string, detail string, thumbnailURL string, startTime time.Time, endTime time.Time, sessionNumber int8, status domain.LiveStatus) error
	deleteFn  func(ctx context.Context, id int) error
	getByIDFn func(ctx context.Context, id int) (*domain.Live, error)
	getAllFn  func(ctx context.Context) ([]*domain.Live, error)
}

func (m *mockLiveUsecase) Create(ctx context.Context, name string, detail string, thumbnailURL string, startTime time.Time, endTime time.Time, sessionNumber int8) error {
	return m.createFn(ctx, name, detail, thumbnailURL, startTime, endTime, sessionNumber)
}
func (m *mockLiveUsecase) Update(ctx context.Context, id int, name string, detail string, thumbnailURL string, startTime time.Time, endTime time.Time, sessionNumber int8, status domain.LiveStatus) error {
	return m.updateFn(ctx, id, name, detail, thumbnailURL, startTime, endTime, sessionNumber, status)
}
func (m *mockLiveUsecase) Delete(ctx context.Context, id int) error {
	return m.deleteFn(ctx, id)
}
func (m *mockLiveUsecase) GetByID(ctx context.Context, id int) (*domain.Live, error) {
	return m.getByIDFn(ctx, id)
}
func (m *mockLiveUsecase) GetAll(ctx context.Context) ([]*domain.Live, error) {
	return m.getAllFn(ctx)
}

func TestLiveHandler_Create(t *testing.T) {
	e := echo.New()

	tests := []struct {
		name           string
		requestJSON    string
		mockCreateErr  error
		expectedStatus int
		expectedBody   string
		wantErr        bool
	}{
		{
			name: "正常系: 201 Created が返る",
			requestJSON: `{
				"name": "テストライブ",
				"detail": "詳細",
				"thumbnail_url": "http://example.com/thumb.png",
				"start_time": "2026-05-29T13:00:00Z",
				"end_time": "2026-05-29T14:00:00Z",
				"session_number": 1
			}`,
			mockCreateErr:  nil,
			expectedStatus: http.StatusCreated,
			wantErr:        false, // ハンドラーはエラーを返さず完了する
		},
		{
			name:          "異常系: JSONのフォーマットが不正（Bindエラー）",
			requestJSON:   `{ invalid json }`,
			mockCreateErr: nil,
			wantErr:       true,
		},
		{
			name: "異常系: Usecase でエラー（カスタムエラーハンドラーに委譲するため return err）",
			requestJSON: `{
				"name": "テストライブ",
				"session_number": 1
			}`,
			mockCreateErr:  errors.New("usecase error"), // ここではどんなエラーでも上に投げるかテストする
			expectedStatus: 0,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// リクエストの作成
			req := httptest.NewRequest(http.MethodPost, "/lives", strings.NewReader(tt.requestJSON))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()

			// Echo v5 の Context 作成
			c := echo.NewContext(req, rec, e)

			// モックの準備
			mockUC := &mockLiveUsecase{
				createFn: func(ctx context.Context, name string, detail string, thumbnailURL string, startTime time.Time, endTime time.Time, sessionNumber int8) error {
					return tt.mockCreateErr
				},
			}

			// ハンドラーの呼び出し
			h := NewLiveHandler(mockUC)
			err := h.Create(c)

			if tt.wantErr {
				require.Error(t, err)
				if tt.mockCreateErr != nil {
					// Usecase で発生したエラーの場合
					assert.Equal(t, tt.mockCreateErr, err)
				} else {
					// mockCreateErr が設定されていないのにエラーになる ＝ Bindエラーの場合
					var he *echo.HTTPError
					require.ErrorAs(t, err, &he) // エラーが *echo.HTTPError 型であることを確認
					assert.Equal(t, http.StatusBadRequest, he.Code) // ステータスコードが400であることを確認
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedStatus, rec.Code)

				if tt.expectedBody != "" {
					// 改行がつくことがあるので Trim する
					assert.Equal(t, tt.expectedBody, strings.TrimSuffix(rec.Body.String(), "\n"))
				}
			}
		})
	}
}

func TestLiveHandler_Update(t *testing.T) {
	e := echo.New()

	tests := []struct {
		name           string
		requestID      string // パスパラメータ用
		requestJSON    string // リクエストボディ
		mockUpdateErr  error
		expectedStatus int
		wantErr        bool
	}{
		{
			name:      "正常系: 204 No Content が返る",
			requestID: "1",
			requestJSON: `{
				"name": "更新ライブ",
				"detail": "詳細",
				"thumbnail_url": "http://example.com/thumb.png",
				"start_time": "2026-05-29T13:00:00Z",
				"end_time": "2026-05-29T14:00:00Z",
				"session_number": 1,
				"status": 1
			}`,
			mockUpdateErr:  nil,
			expectedStatus: http.StatusNoContent,
			wantErr:        false,
		},
		{
			name:          "異常系: JSONのフォーマットが不正（Bindエラー）",
			requestID:     "1",
			requestJSON:   `{ invalid json }`,
			mockUpdateErr: nil,
			wantErr:       true,
		},
		{
			name:      "異常系: Usecase でエラー（そのまま上に投げる）",
			requestID: "1",
			requestJSON: `{
				"name": "更新ライブ"
			}`,
			mockUpdateErr: errors.New("usecase error"),
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(tt.requestJSON))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()

			c := echo.NewContext(req, rec, e)
			// URLパスパラメータ（:id）をテスト内でシミュレートする
			c.SetPath("/lives/:id")
			c.SetPathValues(echo.PathValues{{Name: "id", Value: tt.requestID}})

			mockUC := &mockLiveUsecase{
				updateFn: func(ctx context.Context, id int, name string, detail string, thumbnailURL string, startTime time.Time, endTime time.Time, sessionNumber int8, status domain.LiveStatus) error {
					return tt.mockUpdateErr
				},
			}

			h := NewLiveHandler(mockUC)
			err := h.Update(c)

			if tt.wantErr {
				require.Error(t, err)
				if tt.mockUpdateErr != nil {
					// Usecase で発生したエラーの場合
					assert.Equal(t, tt.mockUpdateErr, err)
				} else {
					// Bindエラーの場合
					var he *echo.HTTPError
					require.ErrorAs(t, err, &he)
					assert.Equal(t, http.StatusBadRequest, he.Code)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedStatus, rec.Code)
			}
		})
	}
}

