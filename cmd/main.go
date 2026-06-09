package main

import (
	"github.com/labstack/echo/v5"

	"github.com/meg44k/KCTFestNav-Backend/internal/handler"
	"github.com/meg44k/KCTFestNav-Backend/internal/repository"
	"github.com/meg44k/KCTFestNav-Backend/internal/router"
	"github.com/meg44k/KCTFestNav-Backend/internal/usecase"
)

func main() {
	e := echo.New()
	e.HTTPErrorHandler = handler.CustomHTTPErrorHandler

	db := 0
	liveRepo := repository.NewLiveRepository(db)
	liveUsecase := usecase.NewLiveUsecase(liveRepo)
	liveHandler := handler.NewLiveHandler(liveUsecase)

	handlers := &handler.Handlers{
		Live: liveHandler,
	}
	router.InitRoutes(e, handlers)

	if err := e.Start(":1323"); err != nil {
		e.Logger.Error("failed to start server", "error", err)
	}
}
