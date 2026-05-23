package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	ocr20191230 "github.com/alibabacloud-go/ocr-20191230/v3/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/spf13/viper"

	"github.com/go-nunu/nunu-layout-advanced/internal/model"
	"github.com/go-nunu/nunu-layout-advanced/internal/repository"
	"github.com/go-nunu/nunu-layout-advanced/pkg/log"
)

type EnterpriseService interface {
	OCR(ctx context.Context, licenseURL string) (*EnterpriseOCRResult, error)
	Create(ctx context.Context, userID int64, input EnterpriseCreateInput) (*model.Enterprise, error)
	Update(ctx context.Context, userID int64, input EnterpriseUpdateInput) error
	Delete(ctx context.Context, userID, id int64) error
	SetDefault(ctx context.Context, userID, id int64) error
	ListByUser(ctx context.Context, userID int64) ([]*model.Enterprise, error)
	ListVerifiedByUser(ctx context.Context, userID int64) ([]*model.Enterprise, error)
	GetByID(ctx context.Context, id int64) (*model.Enterprise, error)
}

// ---- input structs ----

type EnterpriseOCRResult struct {
	Name                string
	SocialCreditCode    string
	LegalRepresentative string
	Address             string
	EstablishedDate     string
	BusinessPeriod      string
	RegisteredCapital   string
	BusinessScope       string
}

type EnterpriseCreateInput struct {
	Name                string
	SocialCreditCode    string
	LegalRepresentative string
	Address             string
	EstablishedDate     string
	BusinessPeriod      string
	RegisteredCapital   string
	BusinessScope       string
	LicenseURL          string
	IsDefault           int
}

type EnterpriseUpdateInput struct {
	ID                  int64
	Name                *string
	LegalRepresentative *string
	Address             *string
	EstablishedDate     *string
	BusinessPeriod      *string
	RegisteredCapital   *string
	BusinessScope       *string
	LicenseURL          *string
	IsDefault           *int
}

var creditCodeRe = regexp.MustCompile(`^[0-9A-Z]{18}$`)

// ---- service impl ----

type enterpriseService struct {
	config             *viper.Viper
	logger             *log.Logger
	tm                 repository.Transaction
	enterpriseRepo     repository.EnterpriseRepository
	ocrClient          *ocr20191230.Client
	mu                 sync.Mutex
}

func NewEnterpriseService(
	config *viper.Viper,
	logger *log.Logger,
	tm repository.Transaction,
	enterpriseRepo repository.EnterpriseRepository,
) EnterpriseService {
	return &enterpriseService{
		config:         config,
		logger:         logger,
		tm:             tm,
		enterpriseRepo: enterpriseRepo,
	}
}

func (s *enterpriseService) Create(ctx context.Context, userID int64, input EnterpriseCreateInput) (*model.Enterprise, error) {
	if !creditCodeRe.MatchString(input.SocialCreditCode) {
		return nil, ErrInvalidCreditCode
	}
	existing, err := s.enterpriseRepo.GetByUserAndCode(ctx, userID, input.SocialCreditCode)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrEnterpriseDuplicate
	}

	now := time.Now()
	e := &model.Enterprise{
		UserID:              userID,
		Name:                input.Name,
		SocialCreditCode:    input.SocialCreditCode,
		LegalRepresentative: input.LegalRepresentative,
		Address:             input.Address,
		EstablishedDate:     input.EstablishedDate,
		BusinessPeriod:      input.BusinessPeriod,
		RegisteredCapital:   input.RegisteredCapital,
		BusinessScope:       input.BusinessScope,
		LicenseURL:          input.LicenseURL,
		IsDefault:           input.IsDefault,
		Status:              model.EnterpriseStatusVerified,
		CreateAt:            now,
		UpdateAt:            now,
	}

	if input.IsDefault == 1 {
		return e, s.tm.Transaction(ctx, func(ctx context.Context) error {
			if err := s.enterpriseRepo.ClearDefault(ctx, userID); err != nil {
				return err
			}
			return s.enterpriseRepo.Create(ctx, e)
		})
	}
	return e, s.enterpriseRepo.Create(ctx, e)
}

func (s *enterpriseService) Update(ctx context.Context, userID int64, input EnterpriseUpdateInput) error {
	e, err := s.enterpriseRepo.GetByID(ctx, input.ID)
	if err != nil {
		return err
	}
	if e == nil {
		return ErrEnterpriseNotFound
	}
	if e.UserID != userID {
		return ErrForbidden
	}

	if input.Name != nil {
		e.Name = *input.Name
	}
	if input.LegalRepresentative != nil {
		e.LegalRepresentative = *input.LegalRepresentative
	}
	if input.Address != nil {
		e.Address = *input.Address
	}
	if input.EstablishedDate != nil {
		e.EstablishedDate = *input.EstablishedDate
	}
	if input.BusinessPeriod != nil {
		e.BusinessPeriod = *input.BusinessPeriod
	}
	if input.RegisteredCapital != nil {
		e.RegisteredCapital = *input.RegisteredCapital
	}
	if input.BusinessScope != nil {
		e.BusinessScope = *input.BusinessScope
	}
	if input.LicenseURL != nil {
		e.LicenseURL = *input.LicenseURL
	}
	e.UpdateAt = time.Now()

	if input.IsDefault != nil && *input.IsDefault == 1 {
		return s.tm.Transaction(ctx, func(ctx context.Context) error {
			if err := s.enterpriseRepo.ClearDefault(ctx, userID); err != nil {
				return err
			}
			e.IsDefault = 1
			return s.enterpriseRepo.Update(ctx, e)
		})
	}
	return s.enterpriseRepo.Update(ctx, e)
}

func (s *enterpriseService) Delete(ctx context.Context, userID, id int64) error {
	e, err := s.enterpriseRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if e == nil {
		return ErrEnterpriseNotFound
	}
	if e.UserID != userID {
		return ErrForbidden
	}
	e.Status = model.EnterpriseStatusDeleted
	e.IsDefault = 0
	e.UpdateAt = time.Now()
	return s.enterpriseRepo.Update(ctx, e)
}

func (s *enterpriseService) SetDefault(ctx context.Context, userID, id int64) error {
	e, err := s.enterpriseRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if e == nil {
		return ErrEnterpriseNotFound
	}
	if e.UserID != userID {
		return ErrForbidden
	}
	if e.Status != model.EnterpriseStatusVerified {
		return ErrEnterpriseNotFound
	}
	return s.tm.Transaction(ctx, func(ctx context.Context) error {
		if err := s.enterpriseRepo.ClearDefault(ctx, userID); err != nil {
			return err
		}
		e.IsDefault = 1
		e.UpdateAt = time.Now()
		return s.enterpriseRepo.Update(ctx, e)
	})
}

func (s *enterpriseService) ListByUser(ctx context.Context, userID int64) ([]*model.Enterprise, error) {
	return s.enterpriseRepo.ListByUser(ctx, userID)
}

func (s *enterpriseService) ListVerifiedByUser(ctx context.Context, userID int64) ([]*model.Enterprise, error) {
	return s.enterpriseRepo.ListVerifiedByUser(ctx, userID)
}

func (s *enterpriseService) GetByID(ctx context.Context, id int64) (*model.Enterprise, error) {
	e, err := s.enterpriseRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if e == nil {
		return nil, ErrEnterpriseNotFound
	}
	return e, nil
}

// ---- OCR ----

func (s *enterpriseService) OCR(ctx context.Context, licenseURL string) (*EnterpriseOCRResult, error) {
	client, err := s.ensureClient()
	if err != nil {
		return nil, fmt.Errorf("init ocr client: %w", err)
	}

	httpResp, err := http.Get(licenseURL) //nolint:noctx
	if err != nil {
		return nil, fmt.Errorf("fetch license image: %w", err)
	}
	defer httpResp.Body.Close()

	req := &ocr20191230.RecognizeBusinessLicenseAdvanceRequest{
		ImageURLObject: httpResp.Body,
	}
	result, err := client.RecognizeBusinessLicenseAdvance(req, &util.RuntimeOptions{})
	if err != nil {
		return nil, fmt.Errorf("ocr recognize: %w", err)
	}
	if result.Body == nil || result.Body.Data == nil {
		return nil, errors.New("empty ocr result")
	}

	d := result.Body.Data
	return &EnterpriseOCRResult{
		Name:                derefString(d.Name),
		SocialCreditCode:    derefString(d.RegisterNumber),
		LegalRepresentative: derefString(d.LegalPerson),
		Address:             derefString(d.Address),
		EstablishedDate:     parseChineseDate(derefString(d.EstablishDate)),
		BusinessPeriod:      parseValidPeriodEnd(derefString(d.ValidPeriod)),
		RegisteredCapital:   derefString(d.Capital),
		BusinessScope:       derefString(d.Business),
	}, nil
}

func (s *enterpriseService) ensureClient() (*ocr20191230.Client, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.ocrClient != nil {
		return s.ocrClient, nil
	}
	ak := s.config.GetString("ocr.access_key_id")
	sk := s.config.GetString("ocr.access_key_secret")
	if ak == "" {
		ak = s.config.GetString("oss.access_key_id")
		sk = s.config.GetString("oss.access_key_secret")
	}
	if ak == "" || sk == "" {
		return nil, errors.New("aliyun credentials not configured (ocr or oss)")
	}
	endpoint := s.config.GetString("ocr.endpoint")
	if endpoint == "" {
		endpoint = "ocr.cn-shanghai.aliyuncs.com"
	}
	cfg := &openapi.Config{
		AccessKeyId:     tea.String(ak),
		AccessKeySecret: tea.String(sk),
		Endpoint:        tea.String(endpoint),
	}
	client, err := ocr20191230.NewClient(cfg)
	if err != nil {
		return nil, err
	}
	s.ocrClient = client
	return client, nil
}

// ---- helpers ----

func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return strings.TrimSpace(*s)
}

var chineseDateRe = regexp.MustCompile(`(\d{4})年(\d{1,2})月(\d{1,2})日`)
var compactDateRe = regexp.MustCompile(`^(\d{4})(\d{2})(\d{2})$`)

func parseChineseDate(s string) string {
	if m := chineseDateRe.FindStringSubmatch(s); len(m) == 4 {
		return fmt.Sprintf("%s-%02s-%02s", m[1], m[2], m[3])
	}
	if m := compactDateRe.FindStringSubmatch(s); len(m) == 4 {
		return fmt.Sprintf("%s-%s-%s", m[1], m[2], m[3])
	}
	return s
}

func parseValidPeriodEnd(s string) string {
	idx := strings.Index(s, "至")
	if idx < 0 {
		return parseChineseDate(s)
	}
	end := strings.TrimSpace(s[idx+len("至"):])
	if end == "长期" || end == "" {
		return "长期"
	}
	return parseChineseDate(end)
}
