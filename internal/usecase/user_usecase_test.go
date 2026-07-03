package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/meg44k/KCTFestNav-Backend/internal/domain"
	"github.com/meg44k/KCTFestNav-Backend/internal/repository"
	"github.com/meg44k/KCTFestNav-Backend/internal/usecase"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

// モック用のRepository
type mockUserRepository struct {
	mockCreate       func(ctx context.Context, user *domain.User) error
	mockGetByID      func(ctx context.Context, id uuid.UUID) (*domain.User, error)
	mockGetByLoginID func(ctx context.Context, loginID string) (*domain.User, error)
	mockGetAll       func(ctx context.Context) ([]*domain.User, error)
	mockUpdate       func(ctx context.Context, user *domain.User) error
	mockDelete       func(ctx context.Context, id uuid.UUID) error
}

func (m *mockUserRepository) Create(ctx context.Context, user *domain.User) error {
	if m.mockCreate != nil {
		return m.mockCreate(ctx, user)
	}
	return nil
}

func (m *mockUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	if m.mockGetByID != nil {
		return m.mockGetByID(ctx, id)
	}
	return nil, nil
}

func (m *mockUserRepository) GetByLoginID(ctx context.Context, loginID string) (*domain.User, error) {
	if m.mockGetByLoginID != nil {
		return m.mockGetByLoginID(ctx, loginID)
	}
	return nil, nil
}

func (m *mockUserRepository) GetAll(ctx context.Context) ([]*domain.User, error) {
	if m.mockGetAll != nil {
		return m.mockGetAll(ctx)
	}
	return nil, nil
}

func (m *mockUserRepository) Update(ctx context.Context, user *domain.User) error {
	if m.mockUpdate != nil {
		return m.mockUpdate(ctx, user)
	}
	return nil
}

func (m *mockUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if m.mockDelete != nil {
		return m.mockDelete(ctx, id)
	}
	return nil
}

func TestUserUsecase_GetByID(t *testing.T) {
	t.Run("正常系: リポジトリからユーザーを取得できること", func(t *testing.T) {
		targetID := uuid.New()
		dummyUser, _ := domain.ReconstructUser(targetID, "テスト", "test", []byte("hash"), 1, domain.RoleStudent)

		mockRepo := &mockUserRepository{
			mockGetByID: func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
				assert.Equal(t, targetID, id)
				return dummyUser, nil
			},
		}
		uc := usecase.NewUserUsecase(mockRepo)

		user, err := uc.GetByID(context.Background(), targetID)
		assert.NoError(t, err)
		assert.Equal(t, dummyUser, user)
	})

	t.Run("異常系: リポジトリがエラーを返した場合はそのままエラーを返すこと", func(t *testing.T) {
		targetID := uuid.New()
		mockRepo := &mockUserRepository{
			mockGetByID: func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
				return nil, errors.New("db error")
			},
		}
		uc := usecase.NewUserUsecase(mockRepo)

		user, err := uc.GetByID(context.Background(), targetID)
		assert.Error(t, err)
		assert.Nil(t, user)
	})
}

func TestUserUsecase_GetMe(t *testing.T) {
	t.Run("正常系: コンテキストからRequestUserを取り出してユーザーを取得できること", func(t *testing.T) {
		targetID := uuid.New()
		dummyUser, _ := domain.ReconstructUser(targetID, "テスト", "test", []byte("hash"), 1, domain.RoleStudent)

		mockRepo := &mockUserRepository{
			mockGetByID: func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
				assert.Equal(t, targetID, id)
				return dummyUser, nil
			},
		}
		uc := usecase.NewUserUsecase(mockRepo)

		reqUser := usecase.RequestUser{ID: targetID, Role: domain.RoleStudent, AssignedBoothID: 1}
		ctx := context.WithValue(context.Background(), usecase.ContextRequestUserKey, reqUser)

		user, err := uc.GetMe(ctx)
		assert.NoError(t, err)
		assert.Equal(t, dummyUser, user)
	})

	t.Run("異常系: コンテキストに情報がない場合は ErrUnauthorized を返すこと", func(t *testing.T) {
		uc := usecase.NewUserUsecase(&mockUserRepository{})
		user, err := uc.GetMe(context.Background())
		assert.ErrorIs(t, err, usecase.ErrUnauthorized)
		assert.Nil(t, user)
	})
}

func TestUserUsecase_Authenticate(t *testing.T) {
	t.Run("正常系: パスワードが一致すればユーザーを返すこと", func(t *testing.T) {
		rawPassword := "mypassword"
		hashed, _ := bcrypt.GenerateFromPassword([]byte(rawPassword), bcrypt.MinCost)
		targetID := uuid.New()
		dummyUser, _ := domain.ReconstructUser(targetID, "テスト", "test", hashed, 1, domain.RoleStudent)

		mockRepo := &mockUserRepository{
			mockGetByLoginID: func(ctx context.Context, loginID string) (*domain.User, error) {
				assert.Equal(t, "test", loginID)
				return dummyUser, nil
			},
		}
		uc := usecase.NewUserUsecase(mockRepo)

		user, err := uc.Authenticate(context.Background(), "test", []byte(rawPassword))
		assert.NoError(t, err)
		assert.Equal(t, dummyUser, user)
	})

	t.Run("異常系: ユーザーが見つからない場合は ErrUnauthorized を返すこと", func(t *testing.T) {
		mockRepo := &mockUserRepository{
			mockGetByLoginID: func(ctx context.Context, loginID string) (*domain.User, error) {
				return nil, repository.ErrNotFound
			},
		}
		uc := usecase.NewUserUsecase(mockRepo)

		user, err := uc.Authenticate(context.Background(), "test", []byte("mypassword"))
		assert.ErrorIs(t, err, usecase.ErrUnauthorized)
		assert.Nil(t, user)
	})

	t.Run("異常系: パスワードが間違っている場合は ErrUnauthorized を返すこと", func(t *testing.T) {
		hashed, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.MinCost)
		targetID := uuid.New()
		dummyUser, _ := domain.ReconstructUser(targetID, "テスト", "test", hashed, 1, domain.RoleStudent)

		mockRepo := &mockUserRepository{
			mockGetByLoginID: func(ctx context.Context, loginID string) (*domain.User, error) {
				return dummyUser, nil
			},
		}
		uc := usecase.NewUserUsecase(mockRepo)

		user, err := uc.Authenticate(context.Background(), "test", []byte("wrongpassword"))
		assert.ErrorIs(t, err, usecase.ErrUnauthorized)
		assert.Nil(t, user)
	})
}

func TestUserUsecase_GetAll(t *testing.T) {
	t.Run("正常系: リポジトリから全てのユーザーを取得できること", func(t *testing.T) {
		dummyUser1, _ := domain.ReconstructUser(uuid.New(), "テスト1", "test1", []byte("hash"), 1, domain.RoleStudent)
		dummyUser2, _ := domain.ReconstructUser(uuid.New(), "テスト2", "test2", []byte("hash"), 2, domain.RoleStudent)
		dummyUsers := []*domain.User{dummyUser1, dummyUser2}

		mockRepo := &mockUserRepository{
			mockGetAll: func(ctx context.Context) ([]*domain.User, error) {
				return dummyUsers, nil
			},
		}
		uc := usecase.NewUserUsecase(mockRepo)

		users, err := uc.GetAll(context.Background())
		assert.NoError(t, err)
		assert.Len(t, users, 2)
		assert.Equal(t, "テスト1", users[0].Name)
		assert.Equal(t, "テスト2", users[1].Name)
	})

	t.Run("異常系: リポジトリがエラーを返した場合はそのままエラーを返すこと", func(t *testing.T) {
		mockRepo := &mockUserRepository{
			mockGetAll: func(ctx context.Context) ([]*domain.User, error) {
				return nil, errors.New("db error")
			},
		}
		uc := usecase.NewUserUsecase(mockRepo)

		users, err := uc.GetAll(context.Background())
		assert.Error(t, err)
		assert.Nil(t, users)
	})
}

func TestUserUsecase_Update(t *testing.T) {
	t.Run("正常系: Admin権限であれば更新できること", func(t *testing.T) {
		mockRepo := &mockUserRepository{
			mockUpdate: func(ctx context.Context, user *domain.User) error {
				assert.Equal(t, "更新後の名前", user.Name)
				assert.Equal(t, "new_login", user.LoginID)
				return nil
			},
		}
		uc := usecase.NewUserUsecase(mockRepo)

		reqUser := usecase.RequestUser{ID: uuid.New(), Role: domain.RoleAdmin, AssignedBoothID: 0}
		ctx := context.WithValue(context.Background(), usecase.ContextRequestUserKey, reqUser)

		err := uc.Update(ctx, uuid.New(), "更新後の名前", "new_login", []byte("pass"), 1, domain.RoleStudent)
		assert.NoError(t, err)
	})

	t.Run("異常系: Admin権限でない場合は ErrForbidden が返ること", func(t *testing.T) {
		uc := usecase.NewUserUsecase(&mockUserRepository{})

		reqUser := usecase.RequestUser{ID: uuid.New(), Role: domain.RoleStudent, AssignedBoothID: 1}
		ctx := context.WithValue(context.Background(), usecase.ContextRequestUserKey, reqUser)

		err := uc.Update(ctx, uuid.New(), "更新後の名前", "new_login", []byte("pass"), 1, domain.RoleStudent)
		assert.ErrorIs(t, err, usecase.ErrForbidden)
	})

	t.Run("異常系: リポジトリがエラーを返した場合はそのままエラーを返すこと", func(t *testing.T) {
		mockRepo := &mockUserRepository{
			mockUpdate: func(ctx context.Context, user *domain.User) error {
				return errors.New("db error")
			},
		}
		uc := usecase.NewUserUsecase(mockRepo)

		reqUser := usecase.RequestUser{ID: uuid.New(), Role: domain.RoleAdmin, AssignedBoothID: 0}
		ctx := context.WithValue(context.Background(), usecase.ContextRequestUserKey, reqUser)

		err := uc.Update(ctx, uuid.New(), "更新後の名前", "new_login", []byte("pass"), 1, domain.RoleStudent)
		assert.Error(t, err)
	})
}
