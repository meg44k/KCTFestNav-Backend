package handler

import (
	"net/http"

	"github.com/labstack/echo/v5"
)

func OK(c *echo.Context) error {
	return c.String(http.StatusOK, "OK")
}
