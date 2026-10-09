package handler

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/meg44k/KCTFestNav-Backend/internal/domain"
	"github.com/meg44k/KCTFestNav-Backend/internal/usecase"
	"golang.org/x/crypto/bcrypt"
)

type Handlers struct {
	Stage        *StageHandler
	Booth        *BoothHandler
	User         *UserHandler
	Announcement *AnnouncementHandler
	Like         *LikeHandler
}

func OK(c *echo.Context) error {
	return c.String(http.StatusOK, "OK")
}

func CustomHTTPErrorHandler(c *echo.Context, err error) {
	code := http.StatusInternalServerError
	message := "internal server error"

	var he *echo.HTTPError
	switch {
	// echo既存機能のエラー(404など)
	case errors.As(err, &he):
		code = he.Code
		message = fmt.Sprintf("%v", he.Message)

	// バリデーションチェック系(Bad request)
	case errors.Is(err, domain.ErrEndTimeAfterStartTime),
		errors.Is(err, domain.ErrNameRequired),
		errors.Is(err, domain.ErrContentRequired),
		errors.Is(err, domain.ErrInvalidDirection),
		errors.Is(err, domain.ErrInvalidFloor),
		errors.Is(err, domain.ErrInvalidCongestion),
		errors.Is(err, domain.ErrNotClassBooth),
		errors.Is(err, domain.ErrInvalidRange),
		errors.Is(err, domain.ErrLoginIDRequired),
		errors.Is(err, domain.ErrCannotDeleteSelf):
		code = http.StatusBadRequest
		message = err.Error()

	// 権限チェック系()
	case errors.Is(err, usecase.ErrForbidden):
		code = http.StatusForbidden
		message = err.Error()
	case errors.Is(err, sql.ErrNoRows):
		code = http.StatusNotFound
		message = "not found"
	case errors.Is(err, usecase.ErrUnauthorized):
		code = http.StatusUnauthorized
		message = "unauthorized"
	case errors.Is(err, domain.ErrLoginIDTaken):
		code = http.StatusConflict
		message = err.Error()
	case errors.Is(err, usecase.ErrTooMany):
		code = http.StatusTooManyRequests
		message = err.Error()

	// パスワード系
	case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
		code = http.StatusUnauthorized
		message = "invalid ID or password"
	}
	c.Logger().Error("HTTP error occurred", "error", err)
	c.JSON(code, map[string]string{"error": message})

}
