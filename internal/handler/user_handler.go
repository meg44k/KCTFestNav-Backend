package handler

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	"github.com/meg44k/KCTFestNav-Backend/internal/auth"
	"github.com/meg44k/KCTFestNav-Backend/internal/domain"
)

type UserUsecase interface {
	Create(
		ctx context.Context,
		name string,
		loginID string,
		password []byte,
		assignedBoothID int,
		role domain.Role,
	) error
	Authenticate(ctx context.Context, LoginID string, Password []byte) (*domain.User, error)
	GetMe(ctx context.Context) (*domain.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	GetAll(ctx context.Context) ([]*domain.User, error)
	Update(
		ctx context.Context,
		id uuid.UUID,
		name string,
		loginID string,
		password []byte,
		assignedBoothID int,
		role domain.Role,
	) error
}

type UserHandler struct {
	userUsecase UserUsecase
	jwtSecret   []byte
}

func NewUserHandler(uc UserUsecase, secret []byte) *UserHandler {
	return &UserHandler{
		userUsecase: uc,
		jwtSecret:   secret,
	}
}

type CreateRequest struct {
	Name            string      `json:"name"`              // ユーザー名
	LoginID         string      `json:"login_id"`          // ログインに用いる文字列
	Password        string      `json:"password"`          // パスワード
	AssignedBoothID int         `json:"assigned_booth_id"` // 配属されたブースのID(1-1の人ならID:1-1など)
	Role            domain.Role `json:"role"`              // 役職 (Admin / Gakuseikai / Student / Member)
}

func (h *UserHandler) Create(c *echo.Context) error {
	var req CreateRequest
	if err := c.Bind(&req); err != nil {
		return err
	}
	if err := h.userUsecase.Create(c.Request().Context(), req.Name, req.LoginID, []byte(req.Password), req.AssignedBoothID, req.Role); err != nil {
		return err
	}
	return c.NoContent(http.StatusCreated)
}

type LoginRequest struct {
	LoginID  string `json:"login_id"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

func (h *UserHandler) Login(c *echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return err
	}
	user, err := h.userUsecase.Authenticate(c.Request().Context(), req.LoginID, []byte(req.Password))
	if err != nil {
		return err
	}

	tokenString, err := auth.GenerateToken(user.ID, user.Role, user.AssignedBoothID, h.jwtSecret)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, LoginResponse{
		Token: tokenString,
	})
}

type GetUserResponse struct {
	ID              uuid.UUID   `json:"id"`
	Name            string      `json:"name"`
	LoginID         string      `json:"login_id"`
	AssignedBoothID int         `json:"assigned_booth_id"`
	Role            domain.Role `json:"role"`
}

func (h *UserHandler) GetMe(c *echo.Context) error {
	user, err := h.userUsecase.GetMe(c.Request().Context())
	if err != nil {
		return err
	}
	res := GetUserResponse{
		ID:              user.ID,
		Name:            user.Name,
		LoginID:         user.LoginID,
		AssignedBoothID: user.AssignedBoothID,
		Role:            user.Role,
	}

	return c.JSON(http.StatusOK, res)
}

func (h *UserHandler) GetByID(c *echo.Context) error {
	id, err := getUUIDParam(c)
	if err != nil {
		return err
	}
	user, err := h.userUsecase.GetByID(c.Request().Context(), id)
	if err != nil {
		return err
	}
	res := GetUserResponse{
		ID:              user.ID,
		Name:            user.Name,
		LoginID:         user.LoginID,
		AssignedBoothID: user.AssignedBoothID,
		Role:            user.Role,
	}
	return c.JSON(http.StatusOK, res)
}

type GetAllUsersResponse struct {
	Users []GetUserResponse `json:"users"`
}

func (h *UserHandler) GetAll(c *echo.Context) error {
	users, err := h.userUsecase.GetAll(c.Request().Context())
	if err != nil {
		return err
	}

	res := make([]GetUserResponse, len(users))
	for i, u := range users {
		res[i] = GetUserResponse{
			ID:              u.ID,
			Name:            u.Name,
			LoginID:         u.LoginID,
			AssignedBoothID: u.AssignedBoothID,
			Role:            u.Role,
		}
	}
	return c.JSON(http.StatusOK, GetAllUsersResponse{
		Users: res,
	})
}

type UpdateUserRequest struct {
	ID              uuid.UUID   `json:"id"`
	Name            string      `json:"name"`
	LoginID         string      `json:"login_id"`
	Password        string      `json:"password"`
	AssignedBoothID int         `json:"assigned_booth_id"`
	Role            domain.Role `json:"role"`
}

func (h *UserHandler) Update(c *echo.Context) error {
	var req UpdateUserRequest
	if err := c.Bind(&req); err != nil {
		return err
	}

	id, err := getUUIDParam(c)
	if err != nil {
		return err
	}

	err = h.userUsecase.Update(
		c.Request().Context(),
		id,
		req.Name,
		req.LoginID,
		[]byte(req.Password),
		req.AssignedBoothID,
		req.Role)
	if err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}
