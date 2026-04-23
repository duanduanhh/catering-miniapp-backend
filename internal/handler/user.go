package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	v1 "github.com/go-nunu/nunu-layout-advanced/api/v1"
	"github.com/go-nunu/nunu-layout-advanced/internal/service"
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
		UserID:                user.ID,
		Avatar:                user.Avatar,
		Name:                  user.Name,
		Sex:                   user.Sex,
		Phone:                 user.Phone,
		ContactVoucherNum:     user.ContactVoucherNum,
		FirstTopStatus:        user.FirstTopStatus,
		NewCustomerStatus:     user.NewCustomerStatus,
		ProfileCompleteStatus: user.ProfileCompleteStatus,
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

// ListInvites godoc
// @Summary 邀请记录
// @Tags 个人中心
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body v1.UserInviteListRequest true "params"
// @Success 200 {object} v1.UserInviteListResponseData
// @Router /user/invites [post]
func (h *UserHandler) ListInvites(ctx *gin.Context) {
	userID := GetUserIdFromCtx(ctx)
	if userID == 0 {
		v1.HandleError(ctx, http.StatusUnauthorized, v1.ErrUnauthorized, v1.ErrUnauthorized.Error())
		return
	}
	var req v1.UserInviteListRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, err.Error())
		return
	}
	list, total, inviteTotal, err := h.userService.ListInvites(ctx, userID, req.PageNum, req.PageSize)
	if err != nil {
		h.logger.WithContext(ctx).Error("userService.ListInvites error", zap.Error(err))
		v1.HandleError(ctx, http.StatusInternalServerError, v1.ErrInternalServerError, err.Error())
		return
	}
	resp := v1.UserInviteListResponseData{
		InviteTotal: inviteTotal,
		List:        make([]v1.UserInviteItem, 0, len(list)),
		Total:       total,
	}
	for _, u := range list {
		resp.List = append(resp.List, v1.UserInviteItem{
			UserID:   u.ID,
			Avatar:   u.Avatar,
			Name:     u.Name,
			CreateAt: formatTime(u.CreateAt),
		})
	}
	v1.HandleSuccess(ctx, resp)
}
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
