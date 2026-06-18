package handler_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
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
