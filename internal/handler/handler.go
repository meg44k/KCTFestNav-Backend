package handler

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strings"

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
	Image        *ImageHandler
	Visit        *VisitHandler
}

func OK(c *echo.Context) error {
	return c.String(http.StatusOK, "OK")
}

func CustomHTTPErrorHandler(c *echo.Context, err error) {
	code := http.StatusInternalServerError
	message := "internal server error"

	var he *echo.HTTPError
	var sc echo.HTTPStatusCoder
	switch {
	// echo既存機能のエラー(404など)
	case errors.As(err, &he):
		code = he.Code
		message = fmt.Sprintf("%v", he.Message)
	// 無い URL(404)・違うメソッド(405)。echo v5 のルーターは *HTTPError ではなく
	// StatusCode() を持つ非公開の型で返すので、ここで拾わないと 500 になる
	case errors.As(err, &sc):
		code = sc.StatusCode()
		message = strings.ToLower(http.StatusText(code))

	// バリデーションチェック系(Bad request)
	case errors.Is(err, domain.ErrEndTimeAfterStartTime),
		errors.Is(err, domain.ErrNameRequired),
		errors.Is(err, domain.ErrContentRequired),
		errors.Is(err, domain.ErrInvalidDirection),
		errors.Is(err, domain.ErrInvalidFloor),
		errors.Is(err, domain.ErrInvalidCongestion),
		errors.Is(err, domain.ErrNotClassBooth),
		errors.Is(err, domain.ErrInvalidRange),
		errors.Is(err, domain.ErrInvalidVisitor),
		errors.Is(err, domain.ErrLoginIDRequired),
		errors.Is(err, domain.ErrCannotDeleteSelf),
		errors.Is(err, domain.ErrInvalidImage),
		errors.Is(err, domain.ErrInvalidImageTarget):
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
	case errors.Is(err, domain.ErrLoginIDTaken), errors.Is(err, domain.ErrImageChanged):
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
