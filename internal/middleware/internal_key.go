package middleware

import (
	"crypto/subtle"
	"net/http"

	"github.com/labstack/echo/v5"
)

// Next.js のサーバーだけが知る合言葉(X-Internal-Key)を確かめる。
// 設定が空なら何も通さない(空と空で一致させない)
func InternalKey(key string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			got := c.Request().Header.Get("X-Internal-Key")
			if key == "" || subtle.ConstantTimeCompare([]byte(got), []byte(key)) != 1 {
				return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
			}
			return next(c)
		}
	}
}
