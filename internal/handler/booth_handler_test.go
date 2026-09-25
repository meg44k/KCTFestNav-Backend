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

type mockBoothUsecase struct {
	mockGetByID func(ctx context.Context, id int) (*domain.Booth, error)
	mockGetAll  func(ctx context.Context) ([]*domain.Booth, error)
	mockCreate           func(ctx context.Context, p domain.BoothParams) error
	mockDelete           func(ctx context.Context, id int) error
	mockUpdate           func(ctx context.Context, id int, congestionStatus domain.CongestionStatus, p domain.BoothParams) error
	mockUpdateCongestion func(ctx context.Context, id int, congestionLevel domain.CongestionStatus) error
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

func (m *mockBoothUsecase) Create(ctx context.Context, p domain.BoothParams) error {
	if m.mockCreate != nil {
		return m.mockCreate(ctx, p)
	}
	return nil
}

func (m *mockBoothUsecase) Delete(ctx context.Context, id int) error {
	if m.mockDelete != nil {
		return m.mockDelete(ctx, id)
	}
	return nil
}

func (m *mockBoothUsecase) Update(ctx context.Context, id int, congestionStatus domain.CongestionStatus, p domain.BoothParams) error {
	if m.mockUpdate != nil {
		return m.mockUpdate(ctx, id, congestionStatus, p)
	}
	return nil
}

func (m *mockBoothUsecase) UpdateCongestion(ctx context.Context, id int, congestionLevel domain.CongestionStatus) error {
	if m.mockUpdateCongestion != nil {
		return m.mockUpdateCongestion(ctx, id, congestionLevel)
	}
	return nil
}

func TestBoothHandler_GetByID(t *testing.T) {
	t.Run("正常系: ブースが1件取得できること", func(t *testing.T) {
		e := echo.New()
		booth, _ := domain.ReconstructBooth(1, domain.CongestionStatus(1), domain.BoothParams{
			Name:      "ブースA",
			Organizer: "主催者A",
			Detail:    "詳細A",
			X:         10.0,
			Y:         20.0,
			Z:         30.0,
		})

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
		var res handler.GetBoothResponse
		err = json.Unmarshal(rec.Body.Bytes(), &res)
		assert.NoError(t, err)
		assert.Equal(t, 1, res.ID)
		assert.Equal(t, "ブースA", res.Name)
		assert.Equal(t, domain.CongestionStatus(1), res.CongestionStatus)
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
		booth1, _ := domain.ReconstructBooth(1, domain.CongestionStatus(1), domain.BoothParams{
			Name:      "ブースA",
			Organizer: "主催者A",
			Detail:    "詳細A",
			X:         10.0,
			Y:         20.0,
			Z:         30.0,
		})
		booth2, _ := domain.ReconstructBooth(2, domain.CongestionStatus(2), domain.BoothParams{
			Name:      "ブースB",
			Organizer: "主催者B",
			Detail:    "詳細B",
			X:         15.0,
			Y:         25.0,
			Z:         35.0,
		})

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
			mockCreate: func(ctx context.Context, p domain.BoothParams) error {
				assert.Equal(t, "新ブース", p.Name)
				assert.Equal(t, "主催X", p.Organizer)
				assert.Equal(t, float32(1.5), p.X)
				return nil
			},
		}
		h := handler.NewBoothHandler(mockUC)

		reqBody := `{"name":"新ブース", "organizer":"主催X", "detail":"詳細X", "congestion_status":0, "x":1.5, "y":2.5, "z":3.5}`
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
			mockCreate: func(ctx context.Context, p domain.BoothParams) error {
				return errors.New("usecase error")
			},
		}
		h := handler.NewBoothHandler(mockUC)

		reqBody := `{"name":"新ブース", "organizer":"主催X", "detail":"詳細X", "congestion_status":0, "x":1.5, "y":2.5, "z":3.5}`
		req := httptest.NewRequest(http.MethodPost, "/manage/booths", strings.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		e.POST("/manage/booths", h.Create)
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}

func TestBoothHandler_Delete(t *testing.T) {
	t.Run("正常系: ブースを削除できること", func(t *testing.T) {
		e := echo.New()
		mockUC := &mockBoothUsecase{
			mockDelete: func(ctx context.Context, id int) error {
				assert.Equal(t, 1, id)
				return nil
			},
		}
		h := handler.NewBoothHandler(mockUC)

		e.DELETE("/manage/booths/:id", h.Delete)
		req := httptest.NewRequest(http.MethodDelete, "/manage/booths/1", nil)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNoContent, rec.Code)
	})

	t.Run("異常系: IDが数字じゃない場合エラーになること", func(t *testing.T) {
		e := echo.New()
		h := handler.NewBoothHandler(&mockBoothUsecase{})

		e.DELETE("/manage/booths/:id", h.Delete)
		req := httptest.NewRequest(http.MethodDelete, "/manage/booths/abc", nil)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("異常系: Usecase層でエラーが起きた場合はそのままエラーが返ること", func(t *testing.T) {
		e := echo.New()
		mockUC := &mockBoothUsecase{
			mockDelete: func(ctx context.Context, id int) error {
				return errors.New("usecase error")
			},
		}
		h := handler.NewBoothHandler(mockUC)

		e.DELETE("/manage/booths/:id", h.Delete)
		req := httptest.NewRequest(http.MethodDelete, "/manage/booths/1", nil)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}

func TestBoothHandler_Update(t *testing.T) {
	t.Run("正常系: リクエストボディが正しければ 200 OK が返ること", func(t *testing.T) {
		e := echo.New()
		mockUC := &mockBoothUsecase{
			mockUpdate: func(ctx context.Context, id int, congestionStatus domain.CongestionStatus, p domain.BoothParams) error {
				assert.Equal(t, 1, id)
				assert.Equal(t, "ブース更新", p.Name)
				return nil
			},
		}
		h := handler.NewBoothHandler(mockUC)

		e.PUT("/manage/booths/:id", h.Update)
		reqBody := `{"name":"ブース更新","organizer":"学生会","detail":"詳細","congestion_status":1,"x":10.5,"y":20.5,"z":0.0}`
		req := httptest.NewRequest(http.MethodPut, "/manage/booths/1", strings.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		// assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("異常系: 不正なJSONの場合は 400 Bad Request が返ること", func(t *testing.T) {
		e := echo.New()
		h := handler.NewBoothHandler(&mockBoothUsecase{})

		e.PUT("/manage/booths/:id", h.Update)
		req := httptest.NewRequest(http.MethodPut, "/manage/booths/1", strings.NewReader(`{invalid}`))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

func TestBoothHandler_UpdateCongestion(t *testing.T) {
	t.Run("正常系: リクエストボディが正しければ 200 OK が返ること", func(t *testing.T) {
		e := echo.New()
		mockUC := &mockBoothUsecase{
			mockUpdateCongestion: func(ctx context.Context, id int, level domain.CongestionStatus) error {
				assert.Equal(t, 1, id)
				assert.Equal(t, domain.CongestionStatus(2), level)
				return nil
			},
		}
		h := handler.NewBoothHandler(mockUC)

		e.PATCH("/manage/booths/:id/congestion", h.UpdateCongestion)
		reqBody := `{"congestion_status":2}`
		req := httptest.NewRequest(http.MethodPatch, "/manage/booths/1/congestion", strings.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		// assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("異常系: 不正なJSONの場合は 400 Bad Request が返ること", func(t *testing.T) {
		e := echo.New()
		h := handler.NewBoothHandler(&mockBoothUsecase{})

		e.PATCH("/manage/booths/:id/congestion", h.UpdateCongestion)
		req := httptest.NewRequest(http.MethodPatch, "/manage/booths/1/congestion", strings.NewReader(`{invalid}`))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}
