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
}

func NewUserService(
	service *Service,
	userRepo repository.UserRepository,
	contactVoucherHistoryRepo repository.ContactVoucherHistoryRepository,
) UserService {
	return &userService{
		userRepo:                  userRepo,
		contactVoucherHistoryRepo: contactVoucherHistoryRepo,
		Service:                   service,
	}
}

type userService struct {
	userRepo                  repository.UserRepository
	contactVoucherHistoryRepo repository.ContactVoucherHistoryRepository
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
