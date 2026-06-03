package service

import (
	"context"
	"time"

	"github.com/go-nunu/nunu-layout-advanced/internal/model"
	"github.com/go-nunu/nunu-layout-advanced/internal/repository"
)

type UserService interface {
	GetInfo(ctx context.Context, userID int64) (*model.User, error)
	UpdateInfo(ctx context.Context, userID int64, input UpdateUserInfoInput) error
	UpdateGeo(ctx context.Context, userID int64, input UpdateUserGeoInput) error
	ListInvites(ctx context.Context, userID int64, pageNum, pageSize int) ([]*InviteUserItem, int64, int64, error)
}

func NewUserService(
	service *Service,
	userRepo repository.UserRepository,
	contactVoucherHistoryRepo repository.ContactVoucherHistoryRepository,
	jobRepo repository.JobRepository,
	orderRepo repository.OrderRepository,
) UserService {
	return &userService{
		userRepo:                  userRepo,
		contactVoucherHistoryRepo: contactVoucherHistoryRepo,
		jobRepo:                   jobRepo,
		orderRepo:                 orderRepo,
		Service:                   service,
	}
}

type userService struct {
	userRepo                  repository.UserRepository
	contactVoucherHistoryRepo repository.ContactVoucherHistoryRepository
	jobRepo                   repository.JobRepository
	orderRepo                 repository.OrderRepository
	*Service
}

type UpdateUserInfoInput struct {
	Avatar *string
	Name   *string
	Sex    *int
	Phone  *string
}

type UpdateUserGeoInput struct {
	FirstAreaID  *int
	SecondAreaID *int
	ThirdAreaID  *int
	Address      *string
	Longitude    *float64
	Latitude     *float64
}

func (s *userService) GetInfo(ctx context.Context, userID int64) (*model.User, error) {
	return s.userRepo.GetByID(ctx, userID)
}

func (s *userService) UpdateInfo(ctx context.Context, userID int64, input UpdateUserInfoInput) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if input.Avatar != nil {
		user.Avatar = *input.Avatar
	}
	if input.Name != nil {
		user.Name = *input.Name
	}
	if input.Sex != nil {
		user.Sex = *input.Sex
	}
	if input.Phone != nil {
		user.Phone = *input.Phone
	}
	user.UpdateAt = time.Now()

	// 首次完善个人信息：赠送2张联系券，状态不可回退
	if user.ProfileCompleteStatus == 0 {
		return s.tm.Transaction(ctx, func(ctx context.Context) error {
			user.ProfileCompleteStatus = 1
			if err := s.userRepo.Update(ctx, user); err != nil {
				return err
			}
			lastNum := user.ContactVoucherNum
			nextNum := lastNum + 2
			giftUser := *user
			giftUser.ContactVoucherNum = nextNum
			giftUser.UpdateAt = time.Now()
			if err := s.userRepo.Update(ctx, &giftUser); err != nil {
				return err
			}
			history := &model.ContactVoucherHistory{
				UserID:    userID,
				BizType:   model.ContactVoucherHistoryBuy,
				ChangeNum: 2,
				LastNum:   lastNum,
				NextNum:   nextNum,
				Remark:    "首次完善个人信息赠送",
				CreateAt:  time.Now(),
			}
			return s.contactVoucherHistoryRepo.Create(ctx, history)
		})
	}

	return s.userRepo.Update(ctx, user)
}

func (s *userService) UpdateGeo(ctx context.Context, userID int64, input UpdateUserGeoInput) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if input.FirstAreaID != nil {
		user.FirstAreaID = *input.FirstAreaID
	}
	if input.SecondAreaID != nil {
		user.SecondAreaID = *input.SecondAreaID
	}
	if input.ThirdAreaID != nil {
		user.ThirdAreaID = *input.ThirdAreaID
	}
	if input.Address != nil {
		user.Address = *input.Address
	}
	if input.Longitude != nil {
		user.Longitude = *input.Longitude
	}
	if input.Latitude != nil {
		user.Latitude = *input.Latitude
	}
	user.UpdateAt = time.Now()
	return s.userRepo.Update(ctx, user)
}

type InviteUserItem struct {
	User          *model.User
	LoginStatus   int
	PublishStatus int
	ConsumeStatus int
	VoucherEarned int
}

// ListInvites 返回被邀请人列表、分页总数、邀请总人数（取 user.invite_num）
func (s *userService) ListInvites(ctx context.Context, userID int64, pageNum, pageSize int) ([]*InviteUserItem, int64, int64, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, 0, 0, err
	}
	list, total, err := s.userRepo.ListByInviterID(ctx, userID, pageNum, pageSize)
	if err != nil {
		return nil, 0, 0, err
	}
	if len(list) == 0 {
		return []*InviteUserItem{}, total, int64(user.InviteNum), nil
	}

	userIDs := make([]int64, len(list))
	for i, u := range list {
		userIDs[i] = u.ID
	}
	publishedMap, err := s.jobRepo.HasPublishedByUserIDs(ctx, userIDs)
	if err != nil {
		return nil, 0, 0, err
	}
	consumedMap, err := s.orderRepo.HasPaidOrderByUserIDs(ctx, userIDs)
	if err != nil {
		return nil, 0, 0, err
	}

	const (
		loginVoucher   = 2
		publishVoucher = 3
		consumeVoucher = 5
	)
	items := make([]*InviteUserItem, 0, len(list))
	for _, u := range list {
		publishStatus := 0
		if publishedMap[u.ID] {
			publishStatus = 1
		}
		consumeStatus := 0
		if consumedMap[u.ID] {
			consumeStatus = 1
		}
		items = append(items, &InviteUserItem{
			User:          u,
			LoginStatus:   1,
			PublishStatus: publishStatus,
			ConsumeStatus: consumeStatus,
			VoucherEarned: loginVoucher + publishStatus*publishVoucher + consumeStatus*consumeVoucher,
		})
	}
	return items, total, int64(user.InviteNum), nil
}
