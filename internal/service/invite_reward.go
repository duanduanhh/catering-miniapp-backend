package service

import (
	"context"
	"time"

	"github.com/go-nunu/nunu-layout-advanced/internal/model"
	"github.com/go-nunu/nunu-layout-advanced/internal/repository"
)

// rewardInviter 给被邀请人的邀请人发放联系券奖励。
// inviteeUserID 为被邀请人 ID，函数内部查询其 inviter_id，若为 0 则静默跳过。
func rewardInviter(
	ctx context.Context,
	inviteeUserID int64,
	voucherNum int,
	remark string,
	userRepo repository.UserRepository,
	histRepo repository.ContactVoucherHistoryRepository,
) error {
	invitee, err := userRepo.GetByID(ctx, inviteeUserID)
	if err != nil {
		return err
	}
	if invitee.InviterID == 0 {
		return nil
	}
	inviter, err := userRepo.GetByID(ctx, invitee.InviterID)
	if err != nil {
		return err
	}
	lastNum := inviter.ContactVoucherNum
	nextNum := lastNum + voucherNum
	inviter.ContactVoucherNum = nextNum
	inviter.UpdateAt = time.Now()
	if err := userRepo.Update(ctx, inviter); err != nil {
		return err
	}
	history := &model.ContactVoucherHistory{
		UserID:    invitee.InviterID,
		BizType:   model.ContactVoucherHistoryBuy,
		ChangeNum: voucherNum,
		LastNum:   lastNum,
		NextNum:   nextNum,
		Remark:    remark,
		CreateAt:  time.Now(),
	}
	return histRepo.Create(ctx, history)
}
