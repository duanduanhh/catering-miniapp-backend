package handler

import (
	"github.com/gin-gonic/gin"
	v1 "github.com/go-nunu/nunu-layout-advanced/api/v1"
	"github.com/go-nunu/nunu-layout-advanced/internal/service"
	"go.uber.org/zap"
	"net/http"
)

type UserHandler struct {
	*Handler
	userService service.UserService
}

func NewUserHandler(handler *Handler, userService service.UserService) *UserHandler {
	return &UserHandler{
		Handler:     handler,
		userService: userService,
	}
}

// GetInfo godoc
// @Summary 查询个人信息
// @Tags 用户模块
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} v1.UserInfoResponse
// @Router /user/info [get]
func (h *UserHandler) GetInfo(ctx *gin.Context) {
	userID := GetUserIdFromCtx(ctx)
	if userID == 0 {
		v1.HandleError(ctx, http.StatusUnauthorized, v1.ErrUnauthorized, v1.ErrUnauthorized.Error())
		return
	}
	user, err := h.userService.GetInfo(ctx, userID)
	if err != nil {
		h.logger.WithContext(ctx).Error("userService.GetInfo error", zap.Error(err))
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, err.Error())
		return
	}
	v1.HandleSuccess(ctx, v1.UserInfoResponseData{
		UserID: user.ID,
		Avatar: user.Avatar,
		Name:   user.Name,
		Sex:    user.Sex,
		Phone:  user.Phone,
	})
}

// UpdateGeo godoc
// @Summary 更新位置信息
// @Tags 用户模块
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body v1.UpdateUserGeoRequest true "params"
// @Success 200 {object} v1.Response
// @Router /user/update/geo [post]
func (h *UserHandler) UpdateGeo(ctx *gin.Context) {
	userID := GetUserIdFromCtx(ctx)
	if userID == 0 {
		v1.HandleError(ctx, http.StatusUnauthorized, v1.ErrUnauthorized, v1.ErrUnauthorized.Error())
		return
	}
	var req v1.UpdateUserGeoRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, err.Error())
		return
	}
	input := service.UpdateUserGeoInput{
		FirstAreaID:  req.FirstAreaID,
		SecondAreaID: req.SecondAreaID,
		ThirdAreaID:  req.ThirdAreaID,
		Address:      req.Address,
		Longitude:    req.Longitude,
		Latitude:     req.Latitude,
	}
	if err := h.userService.UpdateGeo(ctx, userID, input); err != nil {
		h.logger.WithContext(ctx).Error("userService.UpdateGeo error", zap.Error(err))
		v1.HandleError(ctx, http.StatusInternalServerError, v1.ErrInternalServerError, err.Error())
		return
	}
	v1.HandleSuccess(ctx, nil)
}

// UpdateInfo godoc
// @Summary 更新个人信息
// @Tags 用户模块
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body v1.UpdateUserInfoRequest true "params"
// @Success 200 {object} v1.Response
// @Router /user/update/info [post]
func (h *UserHandler) UpdateInfo(ctx *gin.Context) {
	userID := GetUserIdFromCtx(ctx)
	if userID == 0 {
		v1.HandleError(ctx, http.StatusUnauthorized, v1.ErrUnauthorized, v1.ErrUnauthorized.Error())
		return
	}
	var req v1.UpdateUserInfoRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, err.Error())
		return
	}
	input := service.UpdateUserInfoInput{
		Avatar: req.Avatar,
		Name:   req.Name,
		Sex:    req.Sex,
		Phone:  req.Phone,
	}
	if err := h.userService.UpdateInfo(ctx, userID, input); err != nil {
		h.logger.WithContext(ctx).Error("userService.UpdateInfo error", zap.Error(err))
		v1.HandleError(ctx, http.StatusInternalServerError, v1.ErrInternalServerError, err.Error())
		return
	}
	v1.HandleSuccess(ctx, nil)
}
