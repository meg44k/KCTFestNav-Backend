package usecase_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"

	"github.com/meg44k/KCTFestNav-Backend/internal/domain"
	"github.com/meg44k/KCTFestNav-Backend/internal/repository"
	"github.com/meg44k/KCTFestNav-Backend/internal/usecase"
)

// ログインの失敗を数える(Redis の代わり)。blocked の ID は試させない
type fakeLoginLimiter struct {
	blocked   map[string]bool
	failed    []string
	succeeded []string
}

func (f *fakeLoginLimiter) Blocked(_ context.Context, loginID string) (bool, error) {
	return f.blocked[loginID], nil
}
func (f *fakeLoginLimiter) Failed(_ context.Context, loginID string) error {
	f.failed = append(f.failed, loginID)
	return nil
}
func (f *fakeLoginLimiter) Succeeded(_ context.Context, loginID string) error {
	f.succeeded = append(f.succeeded, loginID)
	return nil
}

func loginUser(t *testing.T, password string) *domain.User {
	t.Helper()
	hashed, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	u, _ := domain.ReconstructUser(uuid.New(), domain.UserParams{
		Name: "テスト", LoginID: "booth-1", Password: hashed, AssignedBoothID: 1, Role: domain.RoleStudent,
	})
	return u
}

func TestUserUsecase_Authenticate_回数の上限(t *testing.T) {
	t.Run("失敗が上限に達した ID は、正しいパスワードでも断り、パスワードも確かめない", func(t *testing.T) {
		called := false
		repo := &mockUserRepository{mockGetByLoginID: func(context.Context, string) (*domain.User, error) {
			called = true
			return loginUser(t, "right"), nil
		}}
		limiter := &fakeLoginLimiter{blocked: map[string]bool{"booth-1": true}}
		uc := usecase.NewUserUsecase(repo).WithLoginLimiter(limiter)

		_, err := uc.Authenticate(context.Background(), "booth-1", []byte("right"))
		assert.ErrorIs(t, err, usecase.ErrTooMany)
		assert.False(t, called)
	})

	t.Run("パスワード違いと、無い ID の失敗を数える", func(t *testing.T) {
		repo := &mockUserRepository{mockGetByLoginID: func(_ context.Context, id string) (*domain.User, error) {
			if id == "booth-1" {
				return loginUser(t, "right"), nil
			}
			return nil, repository.ErrNotFound
		}}
		limiter := &fakeLoginLimiter{}
		uc := usecase.NewUserUsecase(repo).WithLoginLimiter(limiter)

		_, err := uc.Authenticate(context.Background(), "booth-1", []byte("wrong"))
		assert.ErrorIs(t, err, usecase.ErrUnauthorized)
		_, err = uc.Authenticate(context.Background(), "nobody", []byte("x"))
		assert.ErrorIs(t, err, usecase.ErrUnauthorized)
		assert.Equal(t, []string{"booth-1", "nobody"}, limiter.failed)
	})

	t.Run("ログインできたら、その ID の失敗の数を消す", func(t *testing.T) {
		repo := &mockUserRepository{mockGetByLoginID: func(context.Context, string) (*domain.User, error) {
			return loginUser(t, "right"), nil
		}}
		limiter := &fakeLoginLimiter{}
		uc := usecase.NewUserUsecase(repo).WithLoginLimiter(limiter)

		_, err := uc.Authenticate(context.Background(), "booth-1", []byte("right"))
		assert.NoError(t, err)
		assert.Equal(t, []string{"booth-1"}, limiter.succeeded)
		assert.Empty(t, limiter.failed)
	})
}
