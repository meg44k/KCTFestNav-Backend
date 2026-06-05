package handler

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v5"
)

// パラメータからIDを取得する関数。
// Error共通化のために作った
func getIDParam(c *echo.Context) (int, error) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return 0, echo.NewHTTPError(http.StatusBadRequest, "invalid id format")
	}
	return id, nil
}
