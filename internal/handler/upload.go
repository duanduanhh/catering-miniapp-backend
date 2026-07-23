package handler

import (
	"errors"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	v1 "github.com/go-nunu/nunu-layout-advanced/api/v1"
	"github.com/go-nunu/nunu-layout-advanced/internal/service"
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
// @Param check_mode formData string false "预留审核方式参数；当前微信图片审核已关闭，传入值会被忽略"
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

	result, err := h.uploadService.UploadImage(ctx, localFile, file.Filename, GetUserIdFromCtx(ctx), GetOpenidFromCtx(ctx), ctx.PostForm("check_mode"))
	if err != nil {
		if errors.Is(err, service.ErrImageRiskyContent) {
			v1.HandleError(ctx, http.StatusBadRequest, v1.ErrImageRiskyContent, "图片内容可能违规，请更换后重试")
			return
		}
		if errors.Is(err, service.ErrInvalidImageCheckMode) {
			v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, err.Error())
			return
		}
		h.logger.WithContext(ctx).Error("uploadService.UploadImage error", zap.Error(err))
		v1.HandleError(ctx, http.StatusInternalServerError, v1.ErrInternalServerError, err.Error())
		return
	}
	v1.HandleSuccess(ctx, v1.UploadImageResponseData{
		URL:         result.URL,
		CheckMode:   result.CheckMode,
		AuditStatus: result.AuditStatus,
		TraceID:     result.TraceID,
	})
}

func validateImageFile(file *multipart.FileHeader) error {
	const maxSize = 10 * 1024 * 1024
	if file.Size > maxSize {
		return errors.New("image size exceeds 10MB")
	}
	ext := strings.ToLower(filepath.Ext(file.Filename))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif":
		return nil
	default:
		return errors.New("unsupported image format")
	}
}
