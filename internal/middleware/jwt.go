package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/labstack/echo/v5"
	"github.com/meg44k/KCTFestNav-Backend/internal/auth"
	"github.com/meg44k/KCTFestNav-Backend/internal/usecase"
)

func JWTAuth(secret []byte) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "missing authorization header")
			}

			// Authorization Header: Bearer JWTTOKENPAYLOAD
			parts := strings.Split(authHeader, " ")

			if len(parts) != 2 || parts[0] != "Bearer" {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid authorization format")
			}

			tokenString := parts[1]

			claims, err := auth.ParseToken(tokenString, secret)

			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid or expired token")
			}

			reqUser := usecase.RequestUser{
				ID:              claims.UserID,
				Role:            claims.Role,
				AssignedBoothID: claims.AssignedBoothID,
			}

			ctx := context.WithValue(c.Request().Context(), usecase.ContextRequestUserKey, reqUser)
			c.SetRequest(c.Request().WithContext(ctx))
			return next(c)
		}
	}
}
