package handler

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	v1 "github.com/go-nunu/nunu-layout-advanced/api/v1"
	"github.com/go-nunu/nunu-layout-advanced/internal/service"
	"go.uber.org/zap"
)

type UploadHandler struct {
	*Handler
	uploadService service.UploadService
}

func NewUploadHandler(handler *Handler, uploadService service.UploadService) *UploadHandler {
	return &UploadHandler{
		Handler:       handler,
		uploadService: uploadService,
	}
}

// UploadImage godoc
// @Summary 图片上传
// @Tags 通用接口
// @Accept multipart/form-data
// @Produce json
// @Security Bearer
// @Param file formData file true "file"
// @Success 200 {object} v1.UploadImageResponseData
// @Router /img/upload [post]
func (h *UploadHandler) UploadImage(ctx *gin.Context) {
	file, err := ctx.FormFile("file")
	if err != nil {
		h.logger.WithContext(ctx).Error("upload FormFile error", zap.Error(err))
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, err.Error())
		return
	}
	if err := validateImageFile(file); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, err.Error())
		return
	}
	localFile, err := file.Open()
	if err != nil {
		h.logger.WithContext(ctx).Error("upload Open file error", zap.Error(err))
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, err.Error())
		return
	}
	defer localFile.Close()

	userID := GetUserIdFromCtx(ctx)
	ext := filepath.Ext(file.Filename)
	if ext == "" {
		ext = ".bin"
	}
	filename := fmt.Sprintf("%d_%d%s", userID, time.Now().UnixNano(), ext)
	url, err := h.uploadService.Upload(ctx, localFile, filename)
	if err != nil {
		h.logger.WithContext(ctx).Error("uploadService.Upload error", zap.Error(err))
		v1.HandleError(ctx, http.StatusInternalServerError, v1.ErrInternalServerError, err.Error())
		return
	}
	v1.HandleSuccess(ctx, v1.UploadImageResponseData{URL: url})
}

func validateImageFile(file *multipart.FileHeader) error {
	const maxSize = 20 * 1024 * 1024
	if file.Size > maxSize {
		return fmt.Errorf("image size exceeds 20MB")
	}
	ext := strings.ToLower(filepath.Ext(file.Filename))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".bmp", ".webp":
		return nil
	default:
		return fmt.Errorf("unsupported image format")
	}
}
