package handler

import (
	"io"
	"net/http"

	"github.com/labstack/echo/v5"

	"github.com/meg44k/KCTFestNav-Backend/internal/domain"
	"github.com/meg44k/KCTFestNav-Backend/internal/usecase"
)

type ImageHandler struct {
	uc *usecase.ImageUsecase
}

func NewImageHandler(uc *usecase.ImageUsecase) *ImageHandler {
	return &ImageHandler{uc: uc}
}

// ブース・出演者の写真を上げる。multipart の file と target(booth:<ID> / performer:<ID>)
func (h *ImageHandler) Upload(c *echo.Context) error {
	fh, err := c.FormFile("file")
	if err != nil {
		return domain.ErrInvalidImage
	}
	f, err := fh.Open()
	if err != nil {
		return err
	}
	defer f.Close()
	// 上限を 1 バイト超えて読み、超えていれば CheckImage が弾く
	data, err := io.ReadAll(io.LimitReader(f, 2<<20+1))
	if err != nil {
		return err
	}
	url, err := h.uc.Upload(c.Request().Context(), c.FormValue("target"), data)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, map[string]string{"url": url})
}
