package main

import (
	"github.com/labstack/echo/v5"

	"github.com/meg44k/KCTFestNav-Backend/internal/router"
)

func main() {
	e := echo.New()

	router.InitRoutes(e)

	if err := e.Start(":1323"); err != nil {
		e.Logger.Error("failed to start server", "error", err)
	}
}
