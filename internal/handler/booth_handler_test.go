package handler_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/meg44k/KCTFestNav-Backend/internal/domain"
	"github.com/meg44k/KCTFestNav-Backend/internal/handler"
	"github.com/stretchr/testify/assert"
)

// モック用のBoothUsecase
type mockBoothUsecase struct {
	mockGetByID func(ctx context.Context, id int) (*domain.Booth, error)
	mockGetAll  func(ctx context.Context) ([]*domain.Booth, error)
	mockCreate  func(ctx context.Context, name, organizer, detail string, congestionStatus int8, x, y, z float32) error
}

func (m *mockBoothUsecase) GetByID(ctx context.Context, id int) (*domain.Booth, error) {
	if m.mockGetByID != nil {
		return m.mockGetByID(ctx, id)
	}
	return nil, nil
}

func (m *mockBoothUsecase) GetAll(ctx context.Context) ([]*domain.Booth, error) {
	if m.mockGetAll != nil {
		return m.mockGetAll(ctx)
	}
	return nil, nil
}

func (m *mockBoothUsecase) Create(ctx context.Context, name, organizer, detail string, congestionStatus int8, x, y, z float32) error {
	if m.mockCreate != nil {
		return m.mockCreate(ctx, name, organizer, detail, congestionStatus, x, y, z)
	}
	return nil
}

func TestBoothHandler_GetByID(t *testing.T) {
	t.Run("正常系: ブースが1件取得できること", func(t *testing.T) {
		e := echo.New()
		booth, _ := domain.ReconstructBooth(1, "ブースA", "主催者A", "詳細A", int8(1), 10.0, 20.0, 30.0)

		mockUC := &mockBoothUsecase{
			mockGetByID: func(ctx context.Context, id int) (*domain.Booth, error) {
				assert.Equal(t, 1, id) // 期待通りのIDがUsecaseに渡っているか
				return booth, nil
			},
		}

		h := handler.NewBoothHandler(mockUC)

		e.GET("/booths/:id", h.GetByID)
		req := httptest.NewRequest(http.MethodGet, "/booths/1", nil)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		var err error
		// エラーなく 200 OK が返ること
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		// レスポンスのJSONの中身が正しいか
		var res handler.BoothResponse
		err = json.Unmarshal(rec.Body.Bytes(), &res)
		assert.NoError(t, err)
		assert.Equal(t, 1, res.ID)
		assert.Equal(t, "ブースA", res.Name)
		assert.Equal(t, int8(1), res.CongestionStatus)
	})

	t.Run("異常系: IDが数字じゃない場合エラーになること", func(t *testing.T) {
		e := echo.New()
		h := handler.NewBoothHandler(&mockBoothUsecase{})

		e.GET("/booths/:id", h.GetByID)
		req := httptest.NewRequest(http.MethodGet, "/booths/abc", nil)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code) // エラーが返ること
	})

	t.Run("異常系: Usecaseからエラーが返った場合そのままエラーになること", func(t *testing.T) {
		e := echo.New()
		mockUC := &mockBoothUsecase{
			mockGetByID: func(ctx context.Context, id int) (*domain.Booth, error) {
				return nil, errors.New("db error")
			},
		}

		h := handler.NewBoothHandler(mockUC)

		e.GET("/booths/:id", h.GetByID)
		req := httptest.NewRequest(http.MethodGet, "/booths/1", nil)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}

func TestBoothHandler_GetAll(t *testing.T) {
	t.Run("正常系: 全ブースが取得できること", func(t *testing.T) {
		e := echo.New()
		booth1, _ := domain.ReconstructBooth(1, "ブースA", "主催者A", "詳細A", int8(1), 10.0, 20.0, 30.0)
		booth2, _ := domain.ReconstructBooth(2, "ブースB", "主催者B", "詳細B", int8(2), 15.0, 25.0, 35.0)

		mockUC := &mockBoothUsecase{
			mockGetAll: func(ctx context.Context) ([]*domain.Booth, error) {
				return []*domain.Booth{booth1, booth2}, nil
			},
		}

		h := handler.NewBoothHandler(mockUC)

		e.GET("/booths", h.GetAll)
		req := httptest.NewRequest(http.MethodGet, "/booths", nil)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		var err error

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var res handler.GetAllBoothsResponse
		err = json.Unmarshal(rec.Body.Bytes(), &res)
		assert.NoError(t, err)
		assert.Len(t, res.Booths, 2)
		assert.Equal(t, "ブースA", res.Booths[0].Name)
		assert.Equal(t, "ブースB", res.Booths[1].Name)
	})

	t.Run("異常系: Usecaseからエラーが返った場合そのままエラーになること", func(t *testing.T) {
		e := echo.New()
		mockUC := &mockBoothUsecase{
			mockGetAll: func(ctx context.Context) ([]*domain.Booth, error) {
				return nil, errors.New("db error")
			},
		}

		h := handler.NewBoothHandler(mockUC)

		e.GET("/booths", h.GetAll)
		req := httptest.NewRequest(http.MethodGet, "/booths", nil)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}

func TestBoothHandler_Create(t *testing.T) {
	t.Run("正常系: リクエストボディが正しければ 201 Created が返ること", func(t *testing.T) {
		e := echo.New()
		mockUC := &mockBoothUsecase{
			mockCreate: func(ctx context.Context, name, organizer, detail string, congestionStatus int8, x, y, z float32) error {
				assert.Equal(t, "新ブース", name)
				assert.Equal(t, "主催X", organizer)
				assert.Equal(t, float32(1.5), x)
				return nil
			},
		}
		h := handler.NewBoothHandler(mockUC)

		reqBody := `{"name":"新ブース", "organizer":"主催X", "detail":"詳細X", "congestionStatus":0, "x":1.5, "y":2.5, "z":3.5}`
		req := httptest.NewRequest(http.MethodPost, "/manage/booths", strings.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		e.POST("/manage/booths", h.Create)
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusCreated, rec.Code)
	})

	t.Run("異常系: 不正なJSONの場合は 400 Bad Request が返ること", func(t *testing.T) {
		e := echo.New()
		h := handler.NewBoothHandler(&mockBoothUsecase{})

		reqBody := `{"name":12345}` // nameが文字列じゃない
		req := httptest.NewRequest(http.MethodPost, "/manage/booths", strings.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		e.POST("/manage/booths", h.Create)
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("異常系: Usecase層でエラーが起きた場合はそのままエラーが返ること", func(t *testing.T) {
		e := echo.New()
		mockUC := &mockBoothUsecase{
			mockCreate: func(ctx context.Context, name, organizer, detail string, congestionStatus int8, x, y, z float32) error {
				return errors.New("usecase error")
			},
		}
		h := handler.NewBoothHandler(mockUC)

		reqBody := `{"name":"新ブース", "organizer":"主催X", "detail":"詳細X", "congestionStatus":0, "x":1.5, "y":2.5, "z":3.5}`
		req := httptest.NewRequest(http.MethodPost, "/manage/booths", strings.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		e.POST("/manage/booths", h.Create)
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}
