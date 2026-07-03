package handler

import (
	"net/http"
	"strconv"

	"github.com/google/uuid"
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

func getUUIDParam(c *echo.Context) (uuid.UUID, error) {
	id := c.Param("id")
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return uuid.Nil, echo.NewHTTPError(http.StatusBadRequest, "invalid uuid format")
	}
	return parsedID, nil
}
