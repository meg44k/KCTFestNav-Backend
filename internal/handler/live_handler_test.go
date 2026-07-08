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
	createFn  func(ctx context.Context, p domain.LiveParams) error
	updateFn  func(ctx context.Context, id int, status domain.LiveStatus, p domain.LiveParams) error
	deleteFn  func(ctx context.Context, id int) error
	getByIDFn func(ctx context.Context, id int) (*domain.Live, error)
	getAllFn  func(ctx context.Context) ([]*domain.Live, error)
	getCurrentLiveFn func(ctx context.Context) (*domain.Live, error)
	updateLiveStatusFn func(ctx context.Context, id int, status domain.LiveStatus) error
}

func (m *mockLiveUsecase) Create(ctx context.Context, p domain.LiveParams) error {
	return m.createFn(ctx, p)
}
func (m *mockLiveUsecase) Update(ctx context.Context, id int, status domain.LiveStatus, p domain.LiveParams) error {
	return m.updateFn(ctx, id, status, p)
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
func (m *mockLiveUsecase) GetCurrentLive(ctx context.Context) (*domain.Live, error) {
	if m.getCurrentLiveFn != nil {
		return m.getCurrentLiveFn(ctx)
	}
	return nil, nil
}
func (m *mockLiveUsecase) UpdateLiveStatus(ctx context.Context, id int, status domain.LiveStatus) error {
	return m.updateLiveStatusFn(ctx, id, status)
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
				createFn: func(ctx context.Context, p domain.LiveParams) error {
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
				updateFn: func(ctx context.Context, id int, status domain.LiveStatus, p domain.LiveParams) error {
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

func TestLiveHandler_GetByID(t *testing.T) {
	e := echo.New()

	mockTime := time.Date(2026, 5, 29, 13, 0, 0, 0, time.UTC)
	mockLive, _ := domain.ReconstructLive(1, domain.LiveStatusUpcoming, domain.LiveParams{
		Name:          "ライブA",
		Detail:        "詳細",
		ThumbnailURL:  "http://example.com/thumb.png",
		StartTime:     mockTime,
		EndTime:       mockTime.Add(time.Hour),
		SessionNumber: 1,
	})

	tests := []struct {
		name           string
		requestID      string
		mockGetErr     error
		mockGetResult  *domain.Live
		expectedStatus int
		expectedBody   string
		wantErr        bool
	}{
		{
			name:           "正常系: 200 OK が返る",
			requestID:      "1",
			mockGetErr:     nil,
			mockGetResult:  mockLive,
			expectedStatus: http.StatusOK,
			expectedBody:   `{"id":1,"name":"ライブA","detail":"詳細","thumbnail_url":"http://example.com/thumb.png","start_time":"2026-05-29T13:00:00Z","end_time":"2026-05-29T14:00:00Z","session_number":1,"status":0}`,
			wantErr:        false,
		},
		{
			name:           "異常系: IDが数字ではない（400エラーがセットされる）",
			requestID:      "abc",
			mockGetErr:     nil,
			mockGetResult:  nil,
			expectedStatus: 0,
			wantErr:        true,
		},
		{
			name:           "異常系: Usecase でエラー（そのまま上に投げる）",
			requestID:      "999",
			mockGetErr:     errors.New("not found error"),
			mockGetResult:  nil,
			expectedStatus: 0,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()

			c := echo.NewContext(req, rec, e)
			c.SetPath("/lives/:id")
			c.SetPathValues(echo.PathValues{{Name: "id", Value: tt.requestID}})

			mockUC := &mockLiveUsecase{
				getByIDFn: func(ctx context.Context, id int) (*domain.Live, error) {
					return tt.mockGetResult, tt.mockGetErr
				},
			}

			h := NewLiveHandler(mockUC)
			err := h.GetByID(c)

			if tt.wantErr {
				require.Error(t, err)
				if tt.mockGetErr != nil {
					assert.Equal(t, tt.mockGetErr, err)
				} else {
					var he *echo.HTTPError
					require.ErrorAs(t, err, &he)
					assert.Equal(t, http.StatusBadRequest, he.Code)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedStatus, rec.Code)
				assert.Equal(t, tt.expectedBody, strings.TrimSuffix(rec.Body.String(), "\n"))
			}
		})
	}
}

func TestLiveHandler_GetAll(t *testing.T) {
	e := echo.New()

	mockTime := time.Date(2026, 5, 29, 13, 0, 0, 0, time.UTC)
	mockLive1, _ := domain.ReconstructLive(1, domain.LiveStatusFinished, domain.LiveParams{
		Name:          "ライブA",
		Detail:        "詳細A",
		ThumbnailURL:  "http://example.com/thumb1.png",
		StartTime:     mockTime,
		EndTime:       mockTime.Add(time.Hour),
		SessionNumber: 1,
	})
	mockLive2, _ := domain.ReconstructLive(2, domain.LiveStatusUpcoming, domain.LiveParams{
		Name:          "ライブB",
		Detail:        "詳細B",
		ThumbnailURL:  "http://example.com/thumb2.png",
		StartTime:     mockTime.Add(2 * time.Hour),
		EndTime:       mockTime.Add(3 * time.Hour),
		SessionNumber: 2,
	})

	tests := []struct {
		name           string
		mockGetAllErr  error
		mockGetAllRes  []*domain.Live
		expectedStatus int
		expectedBody   string
		wantErr        bool
	}{
		{
			name:           "正常系: 200 OK と配列が返る",
			mockGetAllErr:  nil,
			mockGetAllRes:  []*domain.Live{mockLive1, mockLive2},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"lives":[{"id":1,"name":"ライブA","detail":"詳細A","thumbnail_url":"http://example.com/thumb1.png","start_time":"2026-05-29T13:00:00Z","end_time":"2026-05-29T14:00:00Z","session_number":1,"status":2},{"id":2,"name":"ライブB","detail":"詳細B","thumbnail_url":"http://example.com/thumb2.png","start_time":"2026-05-29T15:00:00Z","end_time":"2026-05-29T16:00:00Z","session_number":2,"status":0}]}`,
			wantErr:        false,
		},
		{
			name:           "正常系: データが0件の場合は空配列が返る",
			mockGetAllErr:  nil,
			mockGetAllRes:  []*domain.Live{},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"lives":[]}`,
			wantErr:        false,
		},
		{
			name:           "異常系: Usecase でエラー（そのまま上に投げる）",
			mockGetAllErr:  errors.New("db error"),
			mockGetAllRes:  nil,
			expectedStatus: 0,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/lives", nil)
			rec := httptest.NewRecorder()
			c := echo.NewContext(req, rec, e)

			mockUC := &mockLiveUsecase{
				getAllFn: func(ctx context.Context) ([]*domain.Live, error) {
					return tt.mockGetAllRes, tt.mockGetAllErr
				},
			}

			h := NewLiveHandler(mockUC)
			err := h.GetAll(c)

			if tt.wantErr {
				require.Error(t, err)
				assert.Equal(t, tt.mockGetAllErr, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedStatus, rec.Code)
				assert.Equal(t, tt.expectedBody, strings.TrimSuffix(rec.Body.String(), "\n"))
			}
		})
	}
}

func TestLiveHandler_Delete(t *testing.T) {
	e := echo.New()

	tests := []struct {
		name           string
		requestID      string
		mockDeleteErr  error
		expectedStatus int
		wantErr        bool
	}{
		{
			name:           "正常系: 204 No Content が返る",
			requestID:      "1",
			mockDeleteErr:  nil,
			expectedStatus: http.StatusNoContent,
			wantErr:        false,
		},
		{
			name:           "異常系: IDが数字ではない（400エラーがセットされる）",
			requestID:      "abc",
			mockDeleteErr:  nil,
			expectedStatus: 0,
			wantErr:        true,
		},
		{
			name:           "異常系: Usecase でエラー（そのまま上に投げる）",
			requestID:      "999",
			mockDeleteErr:  errors.New("db error"),
			expectedStatus: 0,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodDelete, "/", nil)
			rec := httptest.NewRecorder()

			c := echo.NewContext(req, rec, e)
			c.SetPath("/lives/:id")
			c.SetPathValues(echo.PathValues{{Name: "id", Value: tt.requestID}})

			mockUC := &mockLiveUsecase{
				deleteFn: func(ctx context.Context, id int) error {
					return tt.mockDeleteErr
				},
			}

			h := NewLiveHandler(mockUC)
			err := h.Delete(c)

			if tt.wantErr {
				require.Error(t, err)
				if tt.mockDeleteErr != nil {
					assert.Equal(t, tt.mockDeleteErr, err)
				} else {
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
