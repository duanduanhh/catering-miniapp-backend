package handler

import (
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	v1 "github.com/go-nunu/nunu-layout-advanced/api/v1"
	"go.uber.org/zap"
)

type UploadHandler struct {
	*Handler
}

func NewUploadHandler(handler *Handler) *UploadHandler {
	return &UploadHandler{
		Handler: handler,
	}
}

func (h *UploadHandler) UploadImage(ctx *gin.Context) {
	file, err := ctx.FormFile("file")
	if err != nil {
		h.logger.WithContext(ctx).Error("upload FormFile error", zap.Error(err))
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, err.Error())
		return
	}
	dir := filepath.Join("storage", "uploads")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		h.logger.WithContext(ctx).Error("upload MkdirAll error", zap.Error(err))
		v1.HandleError(ctx, http.StatusInternalServerError, v1.ErrInternalServerError, err.Error())
		return
	}
	filename := time.Now().Format("20060102150405") + "_" + filepath.Base(file.Filename)
	target := filepath.Join(dir, filename)
	if err := ctx.SaveUploadedFile(file, target); err != nil {
		h.logger.WithContext(ctx).Error("upload SaveUploadedFile error", zap.Error(err))
		v1.HandleError(ctx, http.StatusInternalServerError, v1.ErrInternalServerError, err.Error())
		return
	}
	url := "/uploads/" + filename
	v1.HandleSuccess(ctx, v1.UploadImageResponseData{URL: url})
}
