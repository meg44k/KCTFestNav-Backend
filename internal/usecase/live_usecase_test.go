package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/meg44k/KCTFestNav-Backend/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// LiveRepository インターフェースを満たすモック
type mockLiveRepository struct {
	createFn           func(ctx context.Context, live *domain.Live) error
	updateFn           func(ctx context.Context, live *domain.Live) error
	deleteFn           func(ctx context.Context, id int) error
	getByIDFn          func(ctx context.Context, id int) (*domain.Live, error)
	getAllFn           func(ctx context.Context) ([]*domain.Live, error)
	getCurrentLiveFn   func(ctx context.Context) (*domain.Live, error)
	updateLiveStatusFn func(ctx context.Context, id int, status domain.LiveStatus) error
}

func (m *mockLiveRepository) Create(ctx context.Context, live *domain.Live) error {
	return m.createFn(ctx, live)
}

func (m *mockLiveRepository) Update(ctx context.Context, live *domain.Live) error {
	return m.updateFn(ctx, live)
}

func (m *mockLiveRepository) Delete(ctx context.Context, id int) error {
	return m.deleteFn(ctx, id)
}

func (m *mockLiveRepository) GetByID(ctx context.Context, id int) (*domain.Live, error) {
	return m.getByIDFn(ctx, id)
}

func (m *mockLiveRepository) GetAll(ctx context.Context) ([]*domain.Live, error) {
	return m.getAllFn(ctx)
}

func (m *mockLiveRepository) GetCurrentLive(ctx context.Context) (*domain.Live, error) {
	if m.getCurrentLiveFn != nil {
		return m.getCurrentLiveFn(ctx)
	}
	return nil, nil
}

func (m *mockLiveRepository) UpdateLiveStatus(ctx context.Context, id int, status domain.LiveStatus) error {
	return m.updateLiveStatusFn(ctx, id, status)
}

// context に Role を詰めるヘルパー
func ctxWithRole(role domain.Role) context.Context {
	return context.WithValue(context.Background(), ContextUserRoleKey, role)
}

func TestLiveUsecase_Create(t *testing.T) {
	baseTime := time.Date(2026, time.May, 29, 13, 0, 0, 0, time.Local)
	ctx := ctxWithRole(domain.RoleAdmin)
	errDB := errors.New("db error")

	tests := []struct {
		name          string
		inputName     string
		inputDetail   string
		inputThumb    string
		inputStart    time.Time
		inputEnd      time.Time
		sessionNumber int8
		repoErr       error
		expectedErr   error
	}{
		{
			name:          "正常系: ライブを作成できる",
			inputName:     "テストライブ",
			inputDetail:   "説明文",
			inputThumb:    "https://example.com/thumb.png",
			inputStart:    baseTime,
			inputEnd:      baseTime.Add(1 * time.Hour),
			sessionNumber: 1,
			repoErr:       nil,
			expectedErr:   nil,
		},
		{
			name:          "異常系: 名前が空文字",
			inputName:     "",
			inputDetail:   "説明文",
			inputThumb:    "https://example.com/thumb.png",
			inputStart:    baseTime,
			inputEnd:      baseTime.Add(1 * time.Hour),
			sessionNumber: 1,
			repoErr:       nil,
			expectedErr:   domain.ErrNameRequired,
		},
		{
			name:          "異常系: 終了時間が開始時間より前",
			inputName:     "テストライブ",
			inputDetail:   "説明文",
			inputThumb:    "https://example.com/thumb.png",
			inputStart:    baseTime,
			inputEnd:      baseTime.Add(-1 * time.Hour),
			sessionNumber: 1,
			repoErr:       nil,
			expectedErr:   domain.ErrEndTimeAfterStartTime,
		},
		{
			name:          "異常系: sessionNumberが0",
			inputName:     "テストライブ",
			inputDetail:   "説明文",
			inputThumb:    "https://example.com/thumb.png",
			inputStart:    baseTime,
			inputEnd:      baseTime.Add(1 * time.Hour),
			sessionNumber: 0,
			repoErr:       nil,
			expectedErr:   domain.ErrSessionNumberLessThanOne,
		},
		{
			name:          "異常系: DB保存に失敗",
			inputName:     "テストライブ",
			inputDetail:   "説明文",
			inputThumb:    "https://example.com/thumb.png",
			inputStart:    baseTime,
			inputEnd:      baseTime.Add(1 * time.Hour),
			sessionNumber: 1,
			repoErr:       errDB,
			expectedErr:   errDB,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockLiveRepository{
				createFn: func(ctx context.Context, live *domain.Live) error {
					return tt.repoErr
				},
			}

			uc := NewLiveUsecase(repo)
			err := uc.Create(
				ctx,
				tt.inputName,
				tt.inputDetail,
				tt.inputThumb,
				tt.inputStart,
				tt.inputEnd,
				tt.sessionNumber,
			)

			if tt.expectedErr == nil {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedErr)
			}
		})
	}

	t.Run("異常系: 権限不足（Student）", func(t *testing.T) {
		repo := &mockLiveRepository{
			createFn: func(ctx context.Context, live *domain.Live) error { return nil },
		}

		uc := NewLiveUsecase(repo)
		err := uc.Create(
			ctxWithRole(domain.RoleStudent),
			"テストライブ", "説明文", "https://example.com/thumb.png",
			baseTime, baseTime.Add(1*time.Hour), 1,
		)

		require.Error(t, err)
		assert.ErrorIs(t, err, ErrForbidden)
	})

	t.Run("異常系: context に Role がない", func(t *testing.T) {
		repo := &mockLiveRepository{
			createFn: func(ctx context.Context, live *domain.Live) error { return nil },
		}

		uc := NewLiveUsecase(repo)
		err := uc.Create(
			context.Background(),
			"テストライブ", "説明文", "https://example.com/thumb.png",
			baseTime, baseTime.Add(1*time.Hour), 1,
		)

		require.Error(t, err)
		assert.ErrorIs(t, err, ErrForbidden)
	})
}

func TestLiveUsecase_Update(t *testing.T) {
	baseTime := time.Date(2026, time.May, 29, 13, 0, 0, 0, time.Local)
	ctx := ctxWithRole(domain.RoleGakuseikai)
	errDB := errors.New("db error")

	tests := []struct {
		name          string
		inputID       int
		inputName     string
		inputDetail   string
		inputThumb    string
		inputStart    time.Time
		inputEnd      time.Time
		sessionNumber int8
		status        domain.LiveStatus
		repoErr       error
		expectedErr   error
	}{
		{
			name:          "正常系: ライブを更新できる",
			inputID:       1,
			inputName:     "更新ライブ",
			inputDetail:   "更新説明文",
			inputThumb:    "https://example.com/thumb.png",
			inputStart:    baseTime,
			inputEnd:      baseTime.Add(1 * time.Hour),
			sessionNumber: 2,
			status:        domain.LiveStatusOngoing,
			repoErr:       nil,
			expectedErr:   nil,
		},
		{
			name:          "異常系: 名前が空文字",
			inputID:       1,
			inputName:     "",
			inputDetail:   "説明文",
			inputThumb:    "https://example.com/thumb.png",
			inputStart:    baseTime,
			inputEnd:      baseTime.Add(1 * time.Hour),
			sessionNumber: 1,
			status:        domain.LiveStatusUpcoming,
			repoErr:       nil,
			expectedErr:   domain.ErrNameRequired,
		},
		{
			name:          "異常系: 終了時間が開始時間より前",
			inputID:       1,
			inputName:     "テストライブ",
			inputDetail:   "説明文",
			inputThumb:    "https://example.com/thumb.png",
			inputStart:    baseTime,
			inputEnd:      baseTime.Add(-1 * time.Hour),
			sessionNumber: 1,
			status:        domain.LiveStatusUpcoming,
			repoErr:       nil,
			expectedErr:   domain.ErrEndTimeAfterStartTime,
		},
		{
			name:          "異常系: sessionNumberが0",
			inputID:       1,
			inputName:     "テストライブ",
			inputDetail:   "説明文",
			inputThumb:    "https://example.com/thumb.png",
			inputStart:    baseTime,
			inputEnd:      baseTime.Add(1 * time.Hour),
			sessionNumber: 0,
			status:        domain.LiveStatusUpcoming,
			repoErr:       nil,
			expectedErr:   domain.ErrSessionNumberLessThanOne,
		},
		{
			name:          "異常系: DB更新に失敗",
			inputID:       1,
			inputName:     "テストライブ",
			inputDetail:   "説明文",
			inputThumb:    "https://example.com/thumb.png",
			inputStart:    baseTime,
			inputEnd:      baseTime.Add(1 * time.Hour),
			sessionNumber: 1,
			status:        domain.LiveStatusUpcoming,
			repoErr:       errDB,
			expectedErr:   errDB,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockLiveRepository{
				updateFn: func(ctx context.Context, live *domain.Live) error {
					return tt.repoErr
				},
			}

			uc := NewLiveUsecase(repo)
			err := uc.Update(
				ctx,
				tt.inputID,
				tt.inputName,
				tt.inputDetail,
				tt.inputThumb,
				tt.inputStart,
				tt.inputEnd,
				tt.sessionNumber,
				tt.status,
			)

			if tt.expectedErr == nil {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedErr)
			}
		})
	}

	t.Run("異常系: 権限不足（Student）", func(t *testing.T) {
		repo := &mockLiveRepository{
			updateFn: func(ctx context.Context, live *domain.Live) error { return nil },
		}

		uc := NewLiveUsecase(repo)
		err := uc.Update(
			ctxWithRole(domain.RoleStudent),
			1, "テストライブ", "説明文", "https://example.com/thumb.png",
			baseTime, baseTime.Add(1*time.Hour), 1, domain.LiveStatusUpcoming,
		)

		require.Error(t, err)
		assert.ErrorIs(t, err, ErrForbidden)
	})

	t.Run("正常系: Admin でも更新できる", func(t *testing.T) {
		repo := &mockLiveRepository{
			updateFn: func(ctx context.Context, live *domain.Live) error { return nil },
		}

		uc := NewLiveUsecase(repo)
		err := uc.Update(
			ctxWithRole(domain.RoleAdmin),
			1, "テストライブ", "説明文", "https://example.com/thumb.png",
			baseTime, baseTime.Add(1*time.Hour), 1, domain.LiveStatusUpcoming,
		)

		require.NoError(t, err)
	})
}

func TestLiveUsecase_GetByID(t *testing.T) {
	baseTime := time.Date(2026, time.May, 29, 13, 0, 0, 0, time.Local)

	mockLive, _ := domain.ReconstructLive(1, "テストライブ", "説明文", "https://example.com/thumb.png", baseTime, baseTime.Add(1*time.Hour), 1, domain.LiveStatusUpcoming)

	t.Run("正常系: ライブを取得できる", func(t *testing.T) {
		repo := &mockLiveRepository{
			getByIDFn: func(ctx context.Context, id int) (*domain.Live, error) {
				return mockLive, nil
			},
		}

		uc := NewLiveUsecase(repo)
		live, err := uc.GetByID(context.Background(), 1)

		require.NoError(t, err)
		assert.Equal(t, mockLive, live)
	})

	t.Run("異常系: 存在しないID", func(t *testing.T) {
		repo := &mockLiveRepository{
			getByIDFn: func(ctx context.Context, id int) (*domain.Live, error) {
				return nil, errors.New("not found")
			},
		}

		uc := NewLiveUsecase(repo)
		live, err := uc.GetByID(context.Background(), 999)

		require.Error(t, err)
		assert.Nil(t, live)
	})
}

func TestLiveUsecase_GetAll(t *testing.T) {
	baseTime := time.Date(2026, time.May, 29, 13, 0, 0, 0, time.Local)

	mockLive1, _ := domain.ReconstructLive(1, "ライブ1", "説明1", "https://example.com/1.png", baseTime, baseTime.Add(1*time.Hour), 1, domain.LiveStatusUpcoming)
	mockLive2, _ := domain.ReconstructLive(2, "ライブ2", "説明2", "https://example.com/2.png", baseTime.Add(2*time.Hour), baseTime.Add(3*time.Hour), 2, domain.LiveStatusOngoing)

	t.Run("正常系: 複数件取得できる", func(t *testing.T) {
		repo := &mockLiveRepository{
			getAllFn: func(ctx context.Context) ([]*domain.Live, error) {
				return []*domain.Live{mockLive1, mockLive2}, nil
			},
		}

		uc := NewLiveUsecase(repo)
		lives, err := uc.GetAll(context.Background())

		require.NoError(t, err)
		assert.Len(t, lives, 2)
	})

	t.Run("正常系: 0件の場合", func(t *testing.T) {
		repo := &mockLiveRepository{
			getAllFn: func(ctx context.Context) ([]*domain.Live, error) {
				return []*domain.Live{}, nil
			},
		}

		uc := NewLiveUsecase(repo)
		lives, err := uc.GetAll(context.Background())

		require.NoError(t, err)
		assert.Empty(t, lives)
	})

	t.Run("異常系: DB取得に失敗", func(t *testing.T) {
		repo := &mockLiveRepository{
			getAllFn: func(ctx context.Context) ([]*domain.Live, error) {
				return nil, errors.New("db error")
			},
		}

		uc := NewLiveUsecase(repo)
		lives, err := uc.GetAll(context.Background())

		require.Error(t, err)
		assert.Nil(t, lives)
	})
}

func TestLiveUsecase_Delete(t *testing.T) {
	ctx := ctxWithRole(domain.RoleAdmin)

	t.Run("正常系: ライブを削除できる", func(t *testing.T) {
		repo := &mockLiveRepository{
			deleteFn: func(ctx context.Context, id int) error {
				return nil
			},
		}

		uc := NewLiveUsecase(repo)
		err := uc.Delete(ctx, 1)

		require.NoError(t, err)
	})

	t.Run("異常系: DB削除に失敗", func(t *testing.T) {
		repo := &mockLiveRepository{
			deleteFn: func(ctx context.Context, id int) error {
				return errors.New("db error")
			},
		}

		uc := NewLiveUsecase(repo)
		err := uc.Delete(ctx, 1)

		require.Error(t, err)
		assert.EqualError(t, err, "db error")
	})

	t.Run("異常系: 権限不足（Gakuseikai）", func(t *testing.T) {
		repo := &mockLiveRepository{
			deleteFn: func(ctx context.Context, id int) error { return nil },
		}

		uc := NewLiveUsecase(repo)
		err := uc.Delete(ctxWithRole(domain.RoleGakuseikai), 1)

		require.Error(t, err)
		assert.ErrorIs(t, err, ErrForbidden)
	})

	t.Run("異常系: 権限不足（Member）", func(t *testing.T) {
		repo := &mockLiveRepository{
			deleteFn: func(ctx context.Context, id int) error { return nil },
		}

		uc := NewLiveUsecase(repo)
		err := uc.Delete(ctxWithRole(domain.RoleMember), 1)

		require.Error(t, err)
		assert.ErrorIs(t, err, ErrForbidden)
	})
}

func TestLiveUsecase_GetCurrentLive(t *testing.T) {
	t.Run("正常系: 進行中のライブを取得できる", func(t *testing.T) {
		baseTime := time.Date(2026, time.May, 29, 13, 0, 0, 0, time.Local)
		expectedLive, _ := domain.ReconstructLive(1, "テスト", "詳細", "url", baseTime, baseTime.Add(time.Hour), 1, domain.LiveStatus(1))

		repo := &mockLiveRepository{
			getCurrentLiveFn: func(ctx context.Context) (*domain.Live, error) {
				return expectedLive, nil
			},
		}

		uc := NewLiveUsecase(repo)
		live, err := uc.GetCurrentLive(context.Background())

		require.NoError(t, err)
		assert.Equal(t, expectedLive, live)
	})

	t.Run("異常系: リポジトリがエラーを返す", func(t *testing.T) {
		repo := &mockLiveRepository{
			getCurrentLiveFn: func(ctx context.Context) (*domain.Live, error) {
				return nil, errors.New("db error")
			},
		}

		uc := NewLiveUsecase(repo)
		live, err := uc.GetCurrentLive(context.Background())

		require.Error(t, err)
		assert.Nil(t, live)
		assert.EqualError(t, err, "db error")
	})
}

func TestLiveUsecase_UpdateLiveStatus(t *testing.T) {
	t.Run("正常系: 進行中ステータスに更新できる(Admin)", func(t *testing.T) {
		ctx := ctxWithRole(domain.RoleAdmin)

		repo := &mockLiveRepository{
			updateLiveStatusFn: func(ctx context.Context, id int, status domain.LiveStatus) error {
				assert.Equal(t, 1, id)
				assert.Equal(t, domain.LiveStatus(1), status)
				return nil
			},
		}

		uc := NewLiveUsecase(repo)
		err := uc.UpdateLiveStatus(ctx, 1, domain.LiveStatus(1))

		require.NoError(t, err)
	})

	t.Run("異常系: 権限不足(Student)", func(t *testing.T) {
		ctx := ctxWithRole(domain.RoleStudent)
		uc := NewLiveUsecase(&mockLiveRepository{})
		err := uc.UpdateLiveStatus(ctx, 1, domain.LiveStatus(1))

		require.Error(t, err)
		assert.ErrorIs(t, err, ErrForbidden)
	})

	t.Run("異常系: リポジトリがエラーを返す", func(t *testing.T) {
		ctx := ctxWithRole(domain.RoleAdmin)
		repo := &mockLiveRepository{
			updateLiveStatusFn: func(ctx context.Context, id int, status domain.LiveStatus) error {
				return errors.New("db error")
			},
		}

		uc := NewLiveUsecase(repo)
		err := uc.UpdateLiveStatus(ctx, 1, domain.LiveStatus(1))

		require.Error(t, err)
		assert.EqualError(t, err, "db error")
	})
}
