package handler

import (
	"net/http"
	"time"

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
// @Description 返回当前登录用户的基本信息。sex: 0=未设置 1=男 2=女。first_top_status: 0=未享受首次置顶优惠 1=已享受。new_customer_status: 0=未享受新用户优惠 1=已享受。profile_complete_status: 0=未完善 1=已完善（填写手机号）。
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
	shareRefreshAvailable := 1
	if user.ShareRefreshDate != nil {
		now := time.Now()
		ay, am, ad := user.ShareRefreshDate.Date()
		by, bm, bd := now.Date()
		if ay == by && am == bm && ad == bd {
			shareRefreshAvailable = 0
		}
	}
	v1.HandleSuccess(ctx, v1.UserInfoResponseData{
		UserID:                user.ID,
		Avatar:                user.Avatar,
		Name:                  user.Name,
		Sex:                   user.Sex,
		Phone:                 user.Phone,
		UserCode:              user.UserCode,
		ContactVoucherNum:     user.ContactVoucherNum,
		FirstTopStatus:        user.FirstTopStatus,
		NewCustomerStatus:     user.NewCustomerStatus,
		ProfileCompleteStatus: user.ProfileCompleteStatus,
		ShareRefreshAvailable: shareRefreshAvailable,
	})
}

// UpdateGeo godoc
// @Summary 更新位置信息
// @Description 更新用户的地理位置。所有字段均为可选指针，传哪个改哪个。通常在小程序授权地理位置后调用。
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
// @Description 返回当前用户邀请注册的用户列表，按注册时间倒序。invite_total 为累计邀请总人数；total 为当前分页总数。
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
		InviteTotal:          inviteTotal,
		List:                 make([]v1.UserInviteItem, 0, len(list)),
		Total:                total,
		LoginVoucherReward:   2,
		PublishVoucherReward: 3,
		ConsumeVoucherReward: 5,
	}
	for _, u := range list {
		resp.List = append(resp.List, v1.UserInviteItem{
			UserID:        u.User.ID,
			Avatar:        u.User.Avatar,
			Name:          u.User.Name,
			CreateAt:      formatTime(u.User.CreateAt),
			LoginStatus:   u.LoginStatus,
			PublishStatus: u.PublishStatus,
			ConsumeStatus: u.ConsumeStatus,
			VoucherEarned: u.VoucherEarned,
		})
	}
	v1.HandleSuccess(ctx, resp)
}
// UpdateInfo godoc
// @Summary 更新个人信息
// @Description 更新用户基本信息。所有字段均为可选指针，传哪个改哪个。sex: 1=男 2=女。phone 更新后 profile_complete_status 自动置1。
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


