package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/spf13/viper"

	"github.com/go-nunu/nunu-layout-advanced/internal/model"
	"github.com/go-nunu/nunu-layout-advanced/internal/repository"
)

type JobService interface {
	Create(ctx context.Context, userID int64, input JobCreateInput) (*model.Job, error)
	Update(ctx context.Context, userID int64, input JobUpdateInput) error
	Refresh(ctx context.Context, userID, jobID int64) error
	ShareRefresh(ctx context.Context, userID, jobID int64) error
	Close(ctx context.Context, userID, jobID int64, closeReason string) error
	Reopen(ctx context.Context, userID, jobID int64) error
	Delete(ctx context.Context, userID, jobID int64) error
	GetByID(ctx context.Context, jobID int64) (*model.Job, error)
	List(ctx context.Context, query repository.JobListQuery) ([]*model.Job, int64, error)
	ListByUser(ctx context.Context, userID int64, bizType int, status []int, pageNum, pageSize int) ([]*model.Job, int64, error)
	HomeTop(ctx context.Context, bizType, firstAreaID, secondAreaID, limit int) ([]*model.Job, error)
	HomeFeed(ctx context.Context, bizType, firstAreaID, secondAreaID, pageNum, pageSize int) ([]*model.Job, int64, error)
	// PrePublishRent 招租发布（付费）：预建 job(status=待支付) + rent_detail + order，返回支付参数。
	PrePublishRent(ctx context.Context, userID int64, openid string, input RentPrePublishInput) (*RentPrePublishResult, error)
	// GetRentDetailByJobID 招租详情：读取扩展表。
	GetRentDetailByJobID(ctx context.Context, jobID int64) (*model.RentDetail, error)
	// GetRentDetailsByJobIDs 招租列表：按 job_id 批量读扩展表。
	GetRentDetailsByJobIDs(ctx context.Context, ids []int64) (map[int64]*model.RentDetail, error)
	// CleanupPendingRent 招租超时清理：删除超过 ttl 未支付的 job + rent_detail。
	CleanupPendingRent(ctx context.Context, ttl time.Duration, limit int) (int, error)
}

func NewJobService(
	service *Service,
	jobRepository repository.JobRepository,
	userRepository repository.UserRepository,
	contactVoucherHistoryRepository repository.ContactVoucherHistoryRepository,
	rentDetailRepository repository.RentDetailRepository,
	orderRepository repository.OrderRepository,
	orderItemRepository repository.OrderItemRepository,
	payService PayService,
	conf *viper.Viper,
) JobService {
	return &jobService{
		Service:                         service,
		jobRepository:                   jobRepository,
		userRepository:                  userRepository,
		contactVoucherHistoryRepository: contactVoucherHistoryRepository,
		rentDetailRepository:            rentDetailRepository,
		orderRepository:                 orderRepository,
		orderItemRepository:             orderItemRepository,
		payService:                      payService,
		conf:                            conf,
	}
}

type jobService struct {
	*Service
	jobRepository                   repository.JobRepository
	userRepository                  repository.UserRepository
	contactVoucherHistoryRepository repository.ContactVoucherHistoryRepository
	rentDetailRepository            repository.RentDetailRepository
	orderRepository                 repository.OrderRepository
	orderItemRepository             repository.OrderItemRepository
	payService                      PayService
	conf                            *viper.Viper
}

const (
	bizTypeRecruit = 1
	bizTypeResume  = 2
	bizTypeRent    = 3
)

const (
	jobLimitRecruit = 10
	jobLimitResume  = 5
)

// 招租发布相关错误
var (
	ErrRentUseDedicatedAPI = errors.New("rent must be published via /jobs/rent/pre_publish")
	ErrTopNotAllowedForBiz = errors.New("this biz_type does not support top")
	ErrInvalidRentInput    = errors.New("invalid rent input")
)

type JobCreateInput struct {
	BizType            int
	Positions          string
	CompanyName        string
	Longitude          float64
	Latitude           float64
	Address            string
	AddressDetail      string
	Contact            string
	ContanctPersonName string
	Description        string
	PhotoURLs          string
	FirstAreaID        int
	FirstAreaDes       string
	SecondAreaID       int
	SecondAreaDes      string
	ThirdAreaID        int
	ThirdAreaDes       string
	FourAreaID         int
	FourAreaDes        string
	SalaryMin          int
	SalaryMax          int
	BasicProtection    string
	SalaryBenefits     string
	AttendanceLeave    string
	EnterpriseID       int64
	RecruitNum         int
	WorkContent        string
}

type JobUpdateInput struct {
	ID                int64
	Positions         *string
	CompanyName       *string
	ContactPersonName *string
	Longitude         *float64
	Latitude          *float64
	Address           *string
	AddressDetail     *string
	Contact           *string
	Description       *string
	PhotoURLs         *string
	FirstAreaID       *int
	FirstAreaDes      *string
	SecondAreaID      *int
	SecondAreaDes     *string
	ThirdAreaID       *int
	ThirdAreaDes      *string
	FourAreaID        *int
	FourAreaDes       *string
	SalaryMin         *int
	SalaryMax         *int
	BasicProtection   *string
	SalaryBenefits    *string
	AttendanceLeave   *string
	EnterpriseID      *int64
	RecruitNum        *int
	WorkContent       *string
	// 招租(biz_type=3)扩展字段，均可选（nil=不修改，保留数据库原值）
	MonthlyRent       *int
	AreaSize          *int
	TransferFeeType   *int
	TransferFeeAmount *int
	TransferDesc      *string
}

func (s *jobService) Create(ctx context.Context, userID int64, input JobCreateInput) (*model.Job, error) {
	// 招租业务必须走 /jobs/rent/pre_publish 付费流程，禁止走通用 Create
	if input.BizType == bizTypeRent {
		return nil, ErrRentUseDedicatedAPI
	}
	total, err := s.jobRepository.CountByUser(ctx, userID, input.BizType, model.JobStatusActive)
	if err != nil {
		return nil, err
	}
	limit := jobLimitRecruit
	if input.BizType == bizTypeResume {
		limit = jobLimitResume
	}
	if total >= int64(limit) {
		return nil, ErrJobLimitExceeded
	}

	// 检查是否首次发布（所有非删除状态的岗位数为 0）
	publishedMap, err := s.jobRepository.HasPublishedByUserIDs(ctx, []int64{userID})
	if err != nil {
		return nil, err
	}
	isFirstJob := !publishedMap[userID]
	now := time.Now()
	job := &model.Job{
		UserID:            userID,
		BizType:           input.BizType,
		Positions:         input.Positions,
		CompanyName:       input.CompanyName,
		Longitude:         input.Longitude,
		Latitude:          input.Latitude,
		Address:           input.Address,
		AddressDetail:     input.AddressDetail,
		Contact:           input.Contact,
		ContactPersonName: input.ContanctPersonName,
		Description:       input.Description,
		PhotoURLs:         input.PhotoURLs,
		Status:            model.JobStatusActive,
		FirstAreaID:       input.FirstAreaID,
		FirstAreaDes:      input.FirstAreaDes,
		SecondAreaID:      input.SecondAreaID,
		SecondAreaDes:     input.SecondAreaDes,
		ThirdAreaID:       input.ThirdAreaID,
		ThirdAreaDes:      input.ThirdAreaDes,
		FourAreaID:        input.FourAreaID,
		FourAreaDes:       input.FourAreaDes,
		SalaryMin:         input.SalaryMin,
		SalaryMax:         input.SalaryMax,
		BasicProtection:   input.BasicProtection,
		SalaryBenefits:    input.SalaryBenefits,
		AttendanceLeave:   input.AttendanceLeave,
		EnterpriseID:      input.EnterpriseID,
		RecruitNum:        input.RecruitNum,
		WorkContent:       input.WorkContent,
		CreateAt:          now,
		UpdateAt:          now,
		RefreshTime:       &now,
	}
	if err := s.jobRepository.Create(ctx, job); err != nil {
		return nil, err
	}
	if isFirstJob {
		_ = rewardInviter(ctx, userID, 3, "邀请好友首次发布奖励", s.userRepository, s.contactVoucherHistoryRepository)
	}
	return job, nil
}

func (s *jobService) Update(ctx context.Context, userID int64, input JobUpdateInput) error {
	job, err := s.jobRepository.GetByID(ctx, input.ID)
	if err != nil {
		return err
	}
	if job.UserID != userID {
		return ErrForbidden
	}
	if input.Positions != nil {
		job.Positions = *input.Positions
	}
	if input.CompanyName != nil {
		job.CompanyName = *input.CompanyName
	}
	if input.ContactPersonName != nil {
		job.ContactPersonName = *input.ContactPersonName
	}
	if input.Longitude != nil {
		job.Longitude = *input.Longitude
	}
	if input.Latitude != nil {
		job.Latitude = *input.Latitude
	}
	if input.Address != nil {
		job.Address = *input.Address
	}
	if input.AddressDetail != nil {
		job.AddressDetail = *input.AddressDetail
	}
	if input.Contact != nil {
		job.Contact = *input.Contact
	}
	if input.Description != nil {
		job.Description = *input.Description
	}
	if input.PhotoURLs != nil {
		job.PhotoURLs = *input.PhotoURLs
	}
	if input.FirstAreaID != nil {
		job.FirstAreaID = *input.FirstAreaID
	}
	if input.FirstAreaDes != nil {
		job.FirstAreaDes = *input.FirstAreaDes
	}
	if input.SecondAreaID != nil {
		job.SecondAreaID = *input.SecondAreaID
	}
	if input.SecondAreaDes != nil {
		job.SecondAreaDes = *input.SecondAreaDes
	}
	if input.ThirdAreaID != nil {
		job.ThirdAreaID = *input.ThirdAreaID
	}
	if input.ThirdAreaDes != nil {
		job.ThirdAreaDes = *input.ThirdAreaDes
	}
	if input.FourAreaID != nil {
		job.FourAreaID = *input.FourAreaID
	}
	if input.FourAreaDes != nil {
		job.FourAreaDes = *input.FourAreaDes
	}
	if input.SalaryMin != nil {
		job.SalaryMin = *input.SalaryMin
	}
	if input.SalaryMax != nil {
		job.SalaryMax = *input.SalaryMax
	}
	if input.BasicProtection != nil {
		job.BasicProtection = *input.BasicProtection
	}
	if input.SalaryBenefits != nil {
		job.SalaryBenefits = *input.SalaryBenefits
	}
	if input.AttendanceLeave != nil {
		job.AttendanceLeave = *input.AttendanceLeave
	}
	if input.EnterpriseID != nil {
		job.EnterpriseID = *input.EnterpriseID
	}
	if input.RecruitNum != nil {
		job.RecruitNum = *input.RecruitNum
	}
	if input.WorkContent != nil {
		job.WorkContent = *input.WorkContent
	}
	var rentDetail *model.RentDetail
	hasRentUpdate := input.MonthlyRent != nil || input.AreaSize != nil || input.TransferFeeType != nil ||
		input.TransferFeeAmount != nil || input.TransferDesc != nil
	if hasRentUpdate {
		if job.BizType != bizTypeRent {
			return ErrInvalidRentInput
		}
		detail, err := s.rentDetailRepository.GetByJobID(ctx, job.ID)
		if err != nil {
			return err
		}
		if input.MonthlyRent != nil {
			detail.MonthlyRent = *input.MonthlyRent
		}
		if input.AreaSize != nil {
			detail.AreaSize = *input.AreaSize
		}
		if input.TransferFeeType != nil {
			detail.TransferFeeType = model.TransferFeeType(*input.TransferFeeType)
		}
		if input.TransferFeeAmount != nil {
			detail.TransferFeeAmount = *input.TransferFeeAmount
		}
		if input.TransferDesc != nil {
			detail.TransferDesc = *input.TransferDesc
		}
		if err := validateRentDetail(detail); err != nil {
			return err
		}
		rentDetail = detail
	}
	return s.tm.Transaction(ctx, func(ctx context.Context) error {
		if err := s.jobRepository.Update(ctx, job); err != nil {
			return err
		}
		if rentDetail != nil {
			return s.rentDetailRepository.Update(ctx, rentDetail)
		}
		return nil
	})
}

func validateRentDetail(d *model.RentDetail) error {
	if d.MonthlyRent <= 0 || d.AreaSize <= 0 {
		return ErrInvalidRentInput
	}
	switch d.TransferFeeType {
	case model.TransferFeeFixed:
		if d.TransferFeeAmount <= 0 {
			return ErrInvalidRentInput
		}
	case model.TransferFeeNone, model.TransferFeeNegotiable:
		d.TransferFeeAmount = 0
	default:
		return ErrInvalidRentInput
	}
	return nil
}

func (s *jobService) Refresh(ctx context.Context, userID, jobID int64) error {
	job, err := s.jobRepository.GetByID(ctx, jobID)
	if err != nil {
		return err
	}
	if job.UserID != userID {
		return ErrForbidden
	}
	now := time.Now()
	job.RefreshTime = &now
	return s.jobRepository.Update(ctx, job)
}

func (s *jobService) ShareRefresh(ctx context.Context, userID, jobID int64) error {
	user, err := s.userRepository.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	now := time.Now()
	if user.ShareRefreshDate != nil && sameDay(*user.ShareRefreshDate, now) {
		return ErrShareRefreshLimitExceeded
	}
	job, err := s.jobRepository.GetByID(ctx, jobID)
	if err != nil {
		return err
	}
	if job.UserID != userID {
		return ErrForbidden
	}
	job.RefreshTime = &now
	if err = s.jobRepository.Update(ctx, job); err != nil {
		return err
	}
	user.ShareRefreshDate = &now
	return s.userRepository.Update(ctx, user)
}

func sameDay(a, b time.Time) bool {
	ay, am, ad := a.Date()
	by, bm, bd := b.Date()
	return ay == by && am == bm && ad == bd
}

func (s *jobService) Close(ctx context.Context, userID, jobID int64, closeReason string) error {
	job, err := s.jobRepository.GetByID(ctx, jobID)
	if err != nil {
		return err
	}
	if job.UserID != userID {
		return ErrForbidden
	}
	now := time.Now().Truncate(time.Second)
	job.Status = model.JobStatusUserClosed
	job.CloseReason = closeReason
	job.CloseTime = &now
	job.TopStartTime = nil
	job.TopEndTime = nil
	return s.jobRepository.Update(ctx, job)
}

func (s *jobService) Reopen(ctx context.Context, userID, jobID int64) error {
	job, err := s.jobRepository.GetByID(ctx, jobID)
	if err != nil {
		return err
	}
	if job.UserID != userID {
		return ErrForbidden
	}
	now := time.Now().Truncate(time.Second)
	job.Status = model.JobStatusActive
	job.CloseReason = ""
	job.CloseTime = nil
	job.RefreshTime = &now
	return s.jobRepository.Update(ctx, job)
}

func (s *jobService) Delete(ctx context.Context, userID, jobID int64) error {
	job, err := s.jobRepository.GetByID(ctx, jobID)
	if err != nil {
		return err
	}
	if job.UserID != userID {
		return ErrForbidden
	}
	job.Status = model.JobStatusDeleted
	return s.jobRepository.Update(ctx, job)
}

func (s *jobService) GetByID(ctx context.Context, jobID int64) (*model.Job, error) {
	return s.jobRepository.GetByID(ctx, jobID)
}

func (s *jobService) List(ctx context.Context, query repository.JobListQuery) ([]*model.Job, int64, error) {
	return s.jobRepository.List(ctx, query)
}

func (s *jobService) ListByUser(ctx context.Context, userID int64, bizType int, status []int, pageNum, pageSize int) ([]*model.Job, int64, error) {
	return s.jobRepository.ListByUser(ctx, userID, bizType, status, pageNum, pageSize)
}

func (s *jobService) HomeTop(ctx context.Context, bizType, firstAreaID, secondAreaID, limit int) ([]*model.Job, error) {
	return s.jobRepository.ListTop(ctx, bizType, firstAreaID, secondAreaID, limit)
}

func (s *jobService) HomeFeed(ctx context.Context, bizType, firstAreaID, secondAreaID, pageNum, pageSize int) ([]*model.Job, int64, error) {
	return s.jobRepository.ListFeed(ctx, bizType, firstAreaID, secondAreaID, pageNum, pageSize)
}

// ============ 招租业务（biz_type=3）付费发布 ============

// RentPrePublishInput 招租发布输入：包含通用 job 字段 + rent 扩展字段。
type RentPrePublishInput struct {
	Positions         string // 招租标题
	Address           string
	AddressDetail     string
	Longitude         float64
	Latitude          float64
	Contact           string
	ContactPersonName string
	Description       string
	PhotoURLs         string // CSV
	FirstAreaID       int
	FirstAreaDes      string
	SecondAreaID      int
	SecondAreaDes     string
	ThirdAreaID       int
	ThirdAreaDes      string
	FourAreaID        int
	FourAreaDes       string
	// 招租扩展字段
	MonthlyRent       int                   // 月租(元)
	AreaSize          int                   // 店铺面积(平方米)
	TransferFeeType   model.TransferFeeType // 0=无 1=固定金额 2=面议
	TransferFeeAmount int                   // 转让费金额(元)，仅当 TransferFeeType=1 时有值
	TransferDesc      string                // 转让说明
}

// RentPrePublishResult 招租预发布结果：返回 job/order/支付参数。
type RentPrePublishResult struct {
	JobID     int64
	OrderID   int64
	OrderNo   string
	Amount    float64
	PayParams interface{} // v1.PayParams（避免 service 层依赖 api 包，此处用 interface{}）
}

// PrePublishRent 招租发布（付费）：一次事务预建 job(status=待支付) + rent_detail + order + order_item。
// 支付成功后通过微信回调 → order.HandlePayNotify → jobRepository.ActivatePendingRent 翻转为 Active。
func (s *jobService) PrePublishRent(ctx context.Context, userID int64, openid string, input RentPrePublishInput) (*RentPrePublishResult, error) {
	if openid == "" {
		return nil, errors.New("openid required for rent publish")
	}
	// 基础参数校验
	if input.Positions == "" || input.Address == "" || input.Longitude == 0 || input.Latitude == 0 {
		return nil, ErrInvalidRentInput
	}
	if input.MonthlyRent <= 0 || input.AreaSize <= 0 {
		return nil, ErrInvalidRentInput
	}
	if input.PhotoURLs == "" {
		return nil, ErrInvalidRentInput
	}
	// TransferFeeType=1 时金额必须 > 0；其它类型清零金额
	switch input.TransferFeeType {
	case model.TransferFeeFixed:
		if input.TransferFeeAmount <= 0 {
			return nil, ErrInvalidRentInput
		}
	case model.TransferFeeNone, model.TransferFeeNegotiable:
		input.TransferFeeAmount = 0
	default:
		return nil, ErrInvalidRentInput
	}

	// 读取招租发布定价（元）
	price := s.conf.GetFloat64("rent.publish_price")
	if price <= 0 {
		return nil, errors.New("rent.publish_price is not configured")
	}

	now := time.Now()
	job := &model.Job{
		UserID:            userID,
		BizType:           bizTypeRent,
		Positions:         input.Positions,
		Longitude:         input.Longitude,
		Latitude:          input.Latitude,
		Address:           input.Address,
		AddressDetail:     input.AddressDetail,
		Contact:           input.Contact,
		ContactPersonName: input.ContactPersonName,
		Description:       input.Description,
		PhotoURLs:         input.PhotoURLs,
		Status:            model.JobStatusPendingPay, // 待支付
		FirstAreaID:       input.FirstAreaID,
		FirstAreaDes:      input.FirstAreaDes,
		SecondAreaID:      input.SecondAreaID,
		SecondAreaDes:     input.SecondAreaDes,
		ThirdAreaID:       input.ThirdAreaID,
		ThirdAreaDes:      input.ThirdAreaDes,
		FourAreaID:        input.FourAreaID,
		FourAreaDes:       input.FourAreaDes,
		CreateAt:          now,
		UpdateAt:          now,
	}
	rent := &model.RentDetail{
		MonthlyRent:       input.MonthlyRent,
		AreaSize:          input.AreaSize,
		TransferFeeType:   input.TransferFeeType,
		TransferFeeAmount: input.TransferFeeAmount,
		TransferDesc:      input.TransferDesc,
	}
	order := &model.Order{
		OrderNo:     fmt.Sprintf("RENT%s%d", now.Format("20060102150405"), userID%1000000),
		UserID:      userID,
		AmountTotal: model.NewDecimalFromFloat64(price),
		AmountPaid:  model.NewDecimalFromFloat64(0),
		Currency:    "CNY",
		Status:      model.OrderStatusPending,
		CreateAt:    now,
		UpdateAt:    now,
	}
	item := &model.OrderItem{
		ProductType:       model.ProductTypePublishRent,
		TitleSnapshot:     "发布招租",
		UnitPriceSnapshot: price,
		TargetType:        model.OrderTargetJob,
		CreateAt:          now,
		UpdateAt:          now,
	}

	err := s.tm.Transaction(ctx, func(ctx context.Context) error {
		if err := s.jobRepository.Create(ctx, job); err != nil {
			return err
		}
		rent.JobID = job.ID
		if err := s.rentDetailRepository.Create(ctx, rent); err != nil {
			return err
		}
		if err := s.orderRepository.Create(ctx, order); err != nil {
			return err
		}
		item.OrderID = order.ID
		item.TargetID = job.ID
		return s.orderItemRepository.Create(ctx, item)
	})
	if err != nil {
		return nil, err
	}

	// 事务外调用微信支付统一下单，失败不影响 job/order（超时清理任务兜底）
	amountCents, err := order.AmountTotal.ToCents()
	if err != nil {
		return nil, err
	}
	payParams, err := s.payService.BuildPayParams(ctx, order.OrderNo, amountCents, openid, "发布招租")
	if err != nil {
		return nil, err
	}
	return &RentPrePublishResult{
		JobID:     job.ID,
		OrderID:   order.ID,
		OrderNo:   order.OrderNo,
		Amount:    price,
		PayParams: payParams,
	}, nil
}

// GetRentDetailByJobID 招租详情：单条读取扩展表。
func (s *jobService) GetRentDetailByJobID(ctx context.Context, jobID int64) (*model.RentDetail, error) {
	return s.rentDetailRepository.GetByJobID(ctx, jobID)
}

// GetRentDetailsByJobIDs 招租列表：批量读取扩展表。
func (s *jobService) GetRentDetailsByJobIDs(ctx context.Context, ids []int64) (map[int64]*model.RentDetail, error) {
	return s.rentDetailRepository.GetByJobIDs(ctx, ids)
}

// CleanupPendingRent 招租超时清理：删除 create_at < now-ttl 且仍处于待支付的 job + rent_detail。
// 定时任务调用；返回本次清理数量。
func (s *jobService) CleanupPendingRent(ctx context.Context, ttl time.Duration, limit int) (int, error) {
	if ttl <= 0 {
		return 0, errors.New("ttl must be positive")
	}
	before := time.Now().Add(-ttl)
	jobs, err := s.jobRepository.ListPendingRentBefore(ctx, before, limit)
	if err != nil {
		return 0, err
	}
	if len(jobs) == 0 {
		return 0, nil
	}
	cleaned := 0
	for _, j := range jobs {
		err := s.tm.Transaction(ctx, func(ctx context.Context) error {
			// 将 job 软删除（status=4），rent_detail 硬删除
			j.Status = model.JobStatusDeleted
			if err := s.jobRepository.Update(ctx, j); err != nil {
				return err
			}
			return s.rentDetailRepository.DeleteByJobID(ctx, j.ID)
		})
		if err != nil {
			continue
		}
		cleaned++
	}
	return cleaned, nil
}
