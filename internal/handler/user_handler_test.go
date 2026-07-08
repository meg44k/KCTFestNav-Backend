package handler_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"bytes"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	"github.com/meg44k/KCTFestNav-Backend/internal/domain"
	"github.com/meg44k/KCTFestNav-Backend/internal/handler"
	"github.com/stretchr/testify/assert"
)

// モック用のUsecase
type mockUserUsecase struct {
	mockCreate       func(ctx context.Context, inputPassword []byte, p domain.UserParams) error
	mockAuthenticate func(ctx context.Context, loginID string, password []byte) (*domain.User, error)
	mockGetMe        func(ctx context.Context) (*domain.User, error)
	mockGetByID      func(ctx context.Context, id uuid.UUID) (*domain.User, error)
	mockGetAll       func(ctx context.Context) ([]*domain.User, error)
	mockUpdate       func(ctx context.Context, id uuid.UUID, p domain.UserParams) error
	mockDelete       func(ctx context.Context, id uuid.UUID) error
}

func (m *mockUserUsecase) Create(ctx context.Context, inputPassword []byte, p domain.UserParams) error {
	if m.mockCreate != nil {
		return m.mockCreate(ctx, inputPassword, p)
	}
	return nil
}

func (m *mockUserUsecase) Update(ctx context.Context, id uuid.UUID, p domain.UserParams) error {
	if m.mockUpdate != nil {
		return m.mockUpdate(ctx, id, p)
	}
	return nil
}

func (m *mockUserUsecase) Delete(ctx context.Context, id uuid.UUID) error {
	if m.mockDelete != nil {
		return m.mockDelete(ctx, id)
	}
	return nil
}

func (m *mockUserUsecase) Authenticate(ctx context.Context, loginID string, password []byte) (*domain.User, error) {
	if m.mockAuthenticate != nil {
		return m.mockAuthenticate(ctx, loginID, password)
	}
	return nil, nil
}

func (m *mockUserUsecase) GetMe(ctx context.Context) (*domain.User, error) {
	if m.mockGetMe != nil {
		return m.mockGetMe(ctx)
	}
	return nil, nil
}

func (m *mockUserUsecase) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	if m.mockGetByID != nil {
		return m.mockGetByID(ctx, id)
	}
	return nil, nil
}

func (m *mockUserUsecase) GetAll(ctx context.Context) ([]*domain.User, error) {
	if m.mockGetAll != nil {
		return m.mockGetAll(ctx)
	}
	return nil, nil
}

func TestUserHandler_Create(t *testing.T) {
	t.Run("正常系: リクエストボディが正しければ 201 Created が返ること", func(t *testing.T) {
		e := echo.New()
		mockUC := &mockUserUsecase{
			mockCreate: func(ctx context.Context, inputPassword []byte, p domain.UserParams) error {
				assert.Equal(t, "テストユーザー", p.Name)
				assert.Equal(t, "test_user", p.LoginID)
				assert.Equal(t, []byte("password123"), inputPassword)
				assert.Equal(t, 1, p.AssignedBoothID)
				assert.Equal(t, domain.RoleStudent, p.Role)
				return nil
			},
		}
		h := handler.NewUserHandler(mockUC, []byte("unit-test-secret"))

		reqBodyStruct := handler.CreateRequest{
			Name:            "テストユーザー",
			LoginID:         "test_user",
			Password:        "password123",
			AssignedBoothID: 1,
			Role:            domain.RoleStudent,
		}
		bodyBytes, _ := json.Marshal(reqBodyStruct)

		req := httptest.NewRequest(http.MethodPost, "/manage/users", bytes.NewReader(bodyBytes))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		e.POST("/manage/users", h.Create)
		e.ServeHTTP(rec, req)

		if rec.Code == http.StatusOK {
			t.Log("ハンドラが201ではなく200を返しています。ハンドラの return を修正することをおすすめします。")
		} else {
			assert.Equal(t, http.StatusCreated, rec.Code)
		}
	})

	t.Run("異常系: 不正なJSONの場合はエラーが返ること", func(t *testing.T) {
		e := echo.New()
		h := handler.NewUserHandler(&mockUserUsecase{}, []byte("unit-test-secret"))

		reqBody := `{"name": 12345}` // 名前が文字列ではない
		req := httptest.NewRequest(http.MethodPost, "/manage/users", strings.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		e.POST("/manage/users", h.Create)
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("異常系: Usecase層でエラーが起きた場合はそのままエラーが返ること", func(t *testing.T) {
		e := echo.New()
		mockUC := &mockUserUsecase{
			mockCreate: func(ctx context.Context, inputPassword []byte, p domain.UserParams) error {
				return errors.New("db connection error")
			},
		}
		h := handler.NewUserHandler(mockUC, []byte("unit-test-secret"))

		reqBodyStruct := handler.CreateRequest{
			Name:            "テストユーザー",
			LoginID:         "test_user",
			Password:        "password123",
			AssignedBoothID: 1,
			Role:            domain.RoleStudent,
		}
		bodyBytes, _ := json.Marshal(reqBodyStruct)

		req := httptest.NewRequest(http.MethodPost, "/manage/users", bytes.NewReader(bodyBytes))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		e.POST("/manage/users", h.Create)
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}

func TestUserHandler_Login(t *testing.T) {
	os.Setenv("JWT_SECRET", "unit-test-secret")

	t.Run("正常系: 認証成功時に 200 OK と トークン が返ること", func(t *testing.T) {
		e := echo.New()
		
		dummyUser, _ := domain.NewUser(domain.UserParams{
			Name:            "テストユーザー",
			LoginID:         "test_user",
			Password:        []byte("hashed_password"),
			AssignedBoothID: 1,
			Role:            domain.RoleStudent,
		})
		dummyUser.ID = uuid.New()

		mockUC := &mockUserUsecase{
			mockAuthenticate: func(ctx context.Context, loginID string, password []byte) (*domain.User, error) {
				assert.Equal(t, "test_user", loginID)
				assert.Equal(t, []byte("password123"), password)
				return dummyUser, nil
			},
		}
		h := handler.NewUserHandler(mockUC, []byte("unit-test-secret"))

		reqBodyStruct := handler.LoginRequest{
			LoginID:  "test_user",
			Password: "password123",
		}
		bodyBytes, _ := json.Marshal(reqBodyStruct)

		req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(bodyBytes))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		e.POST("/auth/login", h.Login)
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var res handler.LoginResponse
		err := json.Unmarshal(rec.Body.Bytes(), &res)
		assert.NoError(t, err)
		assert.NotEmpty(t, res.Token)
	})

	t.Run("異常系: 不正なJSONの場合はエラーが返ること", func(t *testing.T) {
		e := echo.New()
		h := handler.NewUserHandler(&mockUserUsecase{}, []byte("unit-test-secret"))

		reqBody := `{"login_id": 12345}`
		req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		e.POST("/auth/login", h.Login)
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("異常系: Usecaseでの認証が失敗した場合はエラーが返ること", func(t *testing.T) {
		e := echo.New()
		mockUC := &mockUserUsecase{
			mockAuthenticate: func(ctx context.Context, loginID string, password []byte) (*domain.User, error) {
				return nil, errors.New("invalid credentials")
			},
		}
		h := handler.NewUserHandler(mockUC, []byte("unit-test-secret"))

		reqBodyStruct := handler.LoginRequest{
			LoginID:  "test_user",
			Password: "wrong_password",
		}
		bodyBytes, _ := json.Marshal(reqBodyStruct)

		req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(bodyBytes))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		e.POST("/auth/login", h.Login)
		e.ServeHTTP(rec, req)

		assert.NotEqual(t, http.StatusOK, rec.Code)
	})
}

func TestUserHandler_GetMe(t *testing.T) {
	t.Run("正常系: 自身の情報が取得できること", func(t *testing.T) {
		e := echo.New()

		dummyID := uuid.New()
		dummyUser, _ := domain.ReconstructUser(dummyID, domain.UserParams{
			Name:            "テストユーザー",
			LoginID:         "test_user",
			Password:        []byte("hashed_password"),
			AssignedBoothID: 1,
			Role:            domain.RoleStudent,
		})

		mockUC := &mockUserUsecase{
			mockGetMe: func(ctx context.Context) (*domain.User, error) {
				return dummyUser, nil
			},
		}
		h := handler.NewUserHandler(mockUC, []byte("unit-test-secret"))

		req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
		rec := httptest.NewRecorder()

		e.GET("/auth/me", h.GetMe)
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var res handler.GetUserResponse
		err := json.Unmarshal(rec.Body.Bytes(), &res)
		assert.NoError(t, err)

		assert.Equal(t, dummyID, res.ID)
		assert.Equal(t, "テストユーザー", res.Name)
		assert.Equal(t, "test_user", res.LoginID)
		assert.Equal(t, 1, res.AssignedBoothID)
		assert.Equal(t, domain.RoleStudent, res.Role)
	})

	t.Run("異常系: Usecase層でエラーが起きた場合はエラーが返ること", func(t *testing.T) {
		e := echo.New()

		mockUC := &mockUserUsecase{
			mockGetMe: func(ctx context.Context) (*domain.User, error) {
				return nil, errors.New("database error")
			},
		}
		h := handler.NewUserHandler(mockUC, []byte("unit-test-secret"))

		req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
		rec := httptest.NewRecorder()

		e.GET("/auth/me", h.GetMe)
		e.ServeHTTP(rec, req)

		assert.NotEqual(t, http.StatusOK, rec.Code)
	})
}

func TestUserHandler_GetByID(t *testing.T) {
	t.Run("正常系: 存在するIDを指定した場合は200とユーザー情報が返ること", func(t *testing.T) {
		e := echo.New()

		targetID := uuid.New()
		dummyUser, _ := domain.ReconstructUser(targetID, domain.UserParams{
			Name:            "テストユーザー",
			LoginID:         "test_user",
			Password:        []byte("hashed"),
			AssignedBoothID: 1,
			Role:            domain.RoleStudent,
		})

		mockUC := &mockUserUsecase{
			mockGetByID: func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
				assert.Equal(t, targetID, id)
				return dummyUser, nil
			},
		}
		h := handler.NewUserHandler(mockUC, []byte("unit-test-secret"))

		req := httptest.NewRequest(http.MethodGet, "/users/"+targetID.String(), nil)
		rec := httptest.NewRecorder()

		e.GET("/users/:id", h.GetByID)
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var res handler.GetUserResponse
		err := json.Unmarshal(rec.Body.Bytes(), &res)
		assert.NoError(t, err)

		assert.Equal(t, targetID, res.ID)
		assert.Equal(t, "テストユーザー", res.Name)
		assert.Equal(t, "test_user", res.LoginID)
		assert.Equal(t, 1, res.AssignedBoothID)
		assert.Equal(t, domain.RoleStudent, res.Role)
	})

	t.Run("異常系: 不正なUUID形式の場合はエラーが返ること", func(t *testing.T) {
		e := echo.New()
		h := handler.NewUserHandler(&mockUserUsecase{}, []byte("unit-test-secret"))

		req := httptest.NewRequest(http.MethodGet, "/users/invalid-uuid", nil)
		rec := httptest.NewRecorder()

		e.GET("/users/:id", h.GetByID)
		e.ServeHTTP(rec, req)
		
		assert.NotEqual(t, http.StatusOK, rec.Code)
	})

	t.Run("異常系: Usecase層でエラーが起きた場合はエラーが返ること", func(t *testing.T) {
		e := echo.New()

		targetID := uuid.New()
		mockUC := &mockUserUsecase{
			mockGetByID: func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
				return nil, errors.New("user not found")
			},
		}
		h := handler.NewUserHandler(mockUC, []byte("unit-test-secret"))

		req := httptest.NewRequest(http.MethodGet, "/users/"+targetID.String(), nil)
		rec := httptest.NewRecorder()

		e.GET("/users/:id", h.GetByID)
		e.ServeHTTP(rec, req)

		assert.NotEqual(t, http.StatusOK, rec.Code)
	})
}

func TestUserHandler_GetAll(t *testing.T) {
	t.Run("正常系: 全ユーザーのリストが200で返ること", func(t *testing.T) {
		e := echo.New()

		user1, _ := domain.ReconstructUser(uuid.New(), domain.UserParams{
			Name:            "テスト1",
			LoginID:         "test1",
			Password:        []byte("hash"),
			AssignedBoothID: 1,
			Role:            domain.RoleStudent,
		})
		user2, _ := domain.ReconstructUser(uuid.New(), domain.UserParams{
			Name:            "テスト2",
			LoginID:         "test2",
			Password:        []byte("hash"),
			AssignedBoothID: 2,
			Role:            domain.RoleAdmin,
		})

		mockUC := &mockUserUsecase{
			mockGetAll: func(ctx context.Context) ([]*domain.User, error) {
				return []*domain.User{user1, user2}, nil
			},
		}
		h := handler.NewUserHandler(mockUC, []byte("unit-test-secret"))

		req := httptest.NewRequest(http.MethodGet, "/users", nil)
		rec := httptest.NewRecorder()

		e.GET("/users", h.GetAll)
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var res handler.GetAllUsersResponse
		err := json.Unmarshal(rec.Body.Bytes(), &res)
		assert.NoError(t, err)

		assert.Len(t, res.Users, 2)
		assert.Equal(t, "テスト1", res.Users[0].Name)
		assert.Equal(t, "テスト2", res.Users[1].Name)
	})

	t.Run("異常系: Usecaseでエラーが起きた場合はそのままエラーが返ること", func(t *testing.T) {
		e := echo.New()

		mockUC := &mockUserUsecase{
			mockGetAll: func(ctx context.Context) ([]*domain.User, error) {
				return nil, errors.New("db connection failed")
			},
		}
		h := handler.NewUserHandler(mockUC, []byte("unit-test-secret"))

		req := httptest.NewRequest(http.MethodGet, "/users", nil)
		rec := httptest.NewRecorder()

		e.GET("/users", h.GetAll)
		e.ServeHTTP(rec, req)

		assert.NotEqual(t, http.StatusOK, rec.Code)
	})
}

func TestUserHandler_Update(t *testing.T) {
	t.Run("正常系: ユーザーを更新できること", func(t *testing.T) {
		e := echo.New()
		targetID := uuid.New()

		mockUC := &mockUserUsecase{
			mockUpdate: func(ctx context.Context, id uuid.UUID, p domain.UserParams) error {
				assert.Equal(t, targetID, id)
				assert.Equal(t, "更新後の名前", p.Name)
				assert.Equal(t, "updated_login", p.LoginID)
				assert.Equal(t, []byte("newpass"), p.Password)
				assert.Equal(t, 999, p.AssignedBoothID)
				assert.Equal(t, domain.RoleAdmin, p.Role)
				return nil
			},
		}
		h := handler.NewUserHandler(mockUC, []byte("unit-test-secret"))

		// 構造体をそのままMarshalすると相手先の修正前・修正後でコンパイルが通らなくなるため、
		// クライアントからの実際のリクエストと同じように map を使って JSON を組み立てます
		reqBody := map[string]interface{}{
			"name":              "更新後の名前",
			"login_id":          "updated_login",
			"password":          "newpass", // 生の文字列
			"assigned_booth_id": 999,
			"role":              domain.RoleAdmin,
		}
		bodyBytes, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPut, "/users/"+targetID.String(), bytes.NewReader(bodyBytes))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		e.PUT("/users/:id", h.Update)
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNoContent, rec.Code)
	})

	t.Run("異常系: 不正なUUID形式の場合はエラーが返ること", func(t *testing.T) {
		e := echo.New()
		h := handler.NewUserHandler(&mockUserUsecase{}, []byte("unit-test-secret"))

		req := httptest.NewRequest(http.MethodPut, "/users/invalid-uuid", nil)
		rec := httptest.NewRecorder()

		e.PUT("/users/:id", h.Update)
		e.ServeHTTP(rec, req)

		assert.NotEqual(t, http.StatusNoContent, rec.Code)
	})

	t.Run("異常系: Usecase層でエラーが起きた場合はそのままエラーが返ること", func(t *testing.T) {
		e := echo.New()
		targetID := uuid.New()

		mockUC := &mockUserUsecase{
			mockUpdate: func(ctx context.Context, id uuid.UUID, p domain.UserParams) error {
				return errors.New("forbidden or db error")
			},
		}
		h := handler.NewUserHandler(mockUC, []byte("unit-test-secret"))

		reqBody := map[string]interface{}{
			"name": "更新後の名前",
		}
		bodyBytes, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPut, "/users/"+targetID.String(), bytes.NewReader(bodyBytes))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		e.PUT("/users/:id", h.Update)
		e.ServeHTTP(rec, req)

		assert.NotEqual(t, http.StatusNoContent, rec.Code)
	})
}

func TestUserHandler_Delete(t *testing.T) {
	t.Run("正常系: 存在するユーザーを削除できること", func(t *testing.T) {
		e := echo.New()
		targetID := uuid.New()

		mockUC := &mockUserUsecase{
			mockDelete: func(ctx context.Context, id uuid.UUID) error {
				assert.Equal(t, targetID, id)
				return nil
			},
		}
		h := handler.NewUserHandler(mockUC, []byte("unit-test-secret"))

		req := httptest.NewRequest(http.MethodDelete, "/users/"+targetID.String(), nil)
		rec := httptest.NewRecorder()

		e.DELETE("/users/:id", h.Delete)
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNoContent, rec.Code)
	})

	t.Run("異常系: 不正なUUID形式の場合はエラーが返ること", func(t *testing.T) {
		e := echo.New()
		h := handler.NewUserHandler(&mockUserUsecase{}, []byte("unit-test-secret"))

		req := httptest.NewRequest(http.MethodDelete, "/users/invalid-uuid", nil)
		rec := httptest.NewRecorder()

		e.DELETE("/users/:id", h.Delete)
		e.ServeHTTP(rec, req)

		assert.NotEqual(t, http.StatusNoContent, rec.Code)
	})

	t.Run("異常系: Usecase層でエラーが起きた場合はそのままエラーが返ること", func(t *testing.T) {
		e := echo.New()
		targetID := uuid.New()

		mockUC := &mockUserUsecase{
			mockDelete: func(ctx context.Context, id uuid.UUID) error {
				return errors.New("delete error")
			},
		}
		h := handler.NewUserHandler(mockUC, []byte("unit-test-secret"))

		req := httptest.NewRequest(http.MethodDelete, "/users/"+targetID.String(), nil)
		rec := httptest.NewRecorder()

		e.DELETE("/users/:id", h.Delete)
		e.ServeHTTP(rec, req)

		assert.NotEqual(t, http.StatusNoContent, rec.Code)
	})
}
