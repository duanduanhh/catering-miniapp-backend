package service

import (
	"context"
	"github.com/go-nunu/nunu-layout-advanced/internal/model"
	"github.com/go-nunu/nunu-layout-advanced/internal/repository"
	"time"
)

type UserService interface {
	GetInfo(ctx context.Context, userID int64) (*model.User, error)
	UpdateInfo(ctx context.Context, userID int64, input UpdateUserInfoInput) error
	UpdateGeo(ctx context.Context, userID int64, input UpdateUserGeoInput) error
}

func NewUserService(
	service *Service,
	userRepo repository.UserRepository,
) UserService {
	return &userService{
		userRepo: userRepo,
		Service:  service,
	}
}

type userService struct {
	userRepo repository.UserRepository
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
