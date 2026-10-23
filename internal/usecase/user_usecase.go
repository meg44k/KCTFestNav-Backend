package usecase

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/meg44k/KCTFestNav-Backend/internal/domain"
	"github.com/meg44k/KCTFestNav-Backend/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type UserUsecase struct {
	userRepo domain.UserRepository
	limiter  domain.LoginLimiter
}

// contextにJWTから読み取ったRoleを入れるためのキー
// contextはkey-valueで入ってる
type contextKey string

const ContextRequestUserKey contextKey = "requestUser"

type RequestUser struct {
	ID              uuid.UUID
	Role            domain.Role
	AssignedBoothID int
}

func NewUserUsecase(repo domain.UserRepository) *UserUsecase {
	return &UserUsecase{
		userRepo: repo,
	}
}

// ログインに回数の上限を付ける(付けなければ数えない)
func (uu *UserUsecase) WithLoginLimiter(l domain.LoginLimiter) *UserUsecase {
	uu.limiter = l
	return uu
}

func (uu *UserUsecase) Create(ctx context.Context, inputPassword []byte, p domain.UserParams) error {
	reqUser, ok := ctx.Value(ContextRequestUserKey).(RequestUser)
	if !ok || reqUser.Role != domain.RoleAdmin {
		return ErrForbidden
	}

	hashedPassword, err := bcrypt.GenerateFromPassword(inputPassword, bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	p.Password = hashedPassword
	user, err := domain.NewUser(p)
	if err != nil {
		return err
	}
	return uu.userRepo.Create(ctx, user)
}

// パスワードが違った(または ID が無い)。あと何回試せるか(回数の上限を付けていなければ -1)
type LoginFailedError struct{ Remaining int }

func (e *LoginFailedError) Error() string { return ErrUnauthorized.Error() }
func (e *LoginFailedError) Unwrap() error { return ErrUnauthorized }

// ip はログインした人の接続元(Next.js のサーバーが渡す)。回数の上限を数えるのに使う
func (uu *UserUsecase) Authenticate(ctx context.Context, ip, loginID string, inputPassword []byte) (*domain.User, error) {
	if uu.limiter == nil {
		return uu.authenticate(ctx, loginID, inputPassword)
	}
	// 続けて失敗している接続元は、パスワードを確かめる(重い計算)前に断る
	blocked, err := uu.limiter.Blocked(ctx, ip, loginID)
	if err != nil {
		return nil, err
	}
	if blocked {
		return nil, ErrTooMany
	}

	user, err := uu.authenticate(ctx, loginID, inputPassword)
	switch {
	case errors.Is(err, ErrUnauthorized):
		left, ferr := uu.limiter.Failed(ctx, ip, loginID)
		if ferr != nil {
			return nil, ferr
		}
		if left <= 0 {
			return nil, ErrTooMany
		}
		return nil, &LoginFailedError{Remaining: left}
	case err == nil:
		if serr := uu.limiter.Succeeded(ctx, ip, loginID); serr != nil {
			return nil, serr
		}
	}
	return user, err
}

func (uu *UserUsecase) authenticate(ctx context.Context, loginID string, inputPassword []byte) (*domain.User, error) {
	// データベースからパスワードを取得
	user, err := uu.userRepo.GetByLoginID(ctx, loginID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrUnauthorized
		}
		return nil, err
	}

	// 入力されたパスワードと登録されているパスワードの比較
	if err := bcrypt.CompareHashAndPassword(user.Password, inputPassword); err != nil {
		return nil, ErrUnauthorized
	}
	return user, nil
}

func (uu *UserUsecase) Update(
	ctx context.Context,
	id uuid.UUID,
	p domain.UserParams) error {
	reqUser, ok := ctx.Value(ContextRequestUserKey).(RequestUser)
	if !ok || reqUser.Role != domain.RoleAdmin {
		return ErrForbidden
	}

	p.LoginID = strings.TrimSpace(p.LoginID)
	if p.LoginID == "" {
		return domain.ErrLoginIDRequired
	}

	// パスワードが空なら今のものを保つ。空でなければハッシュ化して保存する
	if len(p.Password) == 0 {
		current, err := uu.userRepo.GetByID(ctx, id)
		if err != nil {
			return err
		}
		p.Password = current.Password
	} else {
		hashed, err := bcrypt.GenerateFromPassword(p.Password, bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		p.Password = hashed
	}

	user, err := domain.ReconstructUser(id, p)
	if err != nil {
		return err
	}
	return uu.userRepo.Update(ctx, user)
}

func (uu *UserUsecase) Delete(ctx context.Context, id uuid.UUID) error {
	reqUser, ok := ctx.Value(ContextRequestUserKey).(RequestUser)
	if !ok || reqUser.Role != domain.RoleAdmin {
		return ErrForbidden
	}
	// 自分を消すとその場で入れなくなる(最後の管理者なら誰も入れなくなる)
	if reqUser.ID == id {
		return domain.ErrCannotDeleteSelf
	}
	return uu.userRepo.Delete(ctx, id)
}

func (uu *UserUsecase) GetMe(ctx context.Context) (*domain.User, error) {
	reqUser, ok := ctx.Value(ContextRequestUserKey).(RequestUser)
	if !ok {
		return nil, ErrUnauthorized
	}
	id := reqUser.ID
	user, err := uu.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (uu *UserUsecase) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	reqUser, ok := ctx.Value(ContextRequestUserKey).(RequestUser)
	if !ok || reqUser.Role != domain.RoleAdmin {
		return nil, ErrForbidden
	}

	user, err := uu.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (uu *UserUsecase) GetAll(ctx context.Context) ([]*domain.User, error) {
	reqUser, ok := ctx.Value(ContextRequestUserKey).(RequestUser)
	if !ok || reqUser.Role != domain.RoleAdmin {
		return nil, ErrForbidden
	}

	users, err := uu.userRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (uu *UserUsecase) InitAdminUser(ctx context.Context, loginID string, inputPassword []byte) error {
	_, err := uu.userRepo.GetByLoginID(ctx, loginID)
	if err == nil {
		return nil
	}
	if !errors.Is(err, repository.ErrNotFound) {
		return err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword(inputPassword, bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user, err := domain.NewUser(domain.UserParams{
		Name:            "admin",
		LoginID:         loginID,
		Password:        hashedPassword,
		AssignedBoothID: 0,
		Role:            domain.RoleAdmin,
	})
	if err != nil {
		return err
	}
	return uu.userRepo.Create(ctx, user)
}
