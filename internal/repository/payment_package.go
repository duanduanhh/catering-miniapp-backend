package repository

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/go-nunu/nunu-layout-advanced/internal/model"
)

type PaymentPackageListQuery struct {
	ProductID *int64
	Status    *int
	Keyword   string
	PageNum   int
	PageSize  int
}

type PaymentPackageRepository interface {
	Create(ctx context.Context, pkg *model.PaymentPackage) error
	GetByID(ctx context.Context, id int64) (*model.PaymentPackage, error)
	GetBySKU(ctx context.Context, skuCode string) (*model.PaymentPackage, error)
	ExistsBySKU(ctx context.Context, skuCode string) (bool, error)
	ExistsByVirtualProductID(ctx context.Context, virtualProductID string, excludeID int64) (bool, error)
	CountByProductID(ctx context.Context, productID int64) (int64, error)
	List(ctx context.Context, query PaymentPackageListQuery) ([]*model.PaymentPackage, int64, error)
	ListAvailable(ctx context.Context, productID int64) ([]*model.PaymentPackage, error)
	Update(ctx context.Context, pkg *model.PaymentPackage, expectedVersion int) (bool, error)
	UpdateStatus(ctx context.Context, id int64, expectedVersion int, from, to model.PaymentPackageStatus, operator string) (bool, error)
	SoftDelete(ctx context.Context, id int64, expectedVersion int, operator string) (bool, error)
	CreateChangeLog(ctx context.Context, log *model.PaymentPackageChangeLog) error
	ListChangeLogs(ctx context.Context, packageID int64, pageNum, pageSize int) ([]*model.PaymentPackageChangeLog, int64, error)
}

func NewPaymentPackageRepository(repository *Repository) PaymentPackageRepository {
	return &paymentPackageRepository{Repository: repository}
}

type paymentPackageRepository struct {
	*Repository
}

func (r *paymentPackageRepository) Create(ctx context.Context, pkg *model.PaymentPackage) error {
	return r.DB(ctx).Create(pkg).Error
}

func (r *paymentPackageRepository) GetByID(ctx context.Context, id int64) (*model.PaymentPackage, error) {
	var pkg model.PaymentPackage
	err := r.DB(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&pkg).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &pkg, nil
}

func (r *paymentPackageRepository) GetBySKU(ctx context.Context, skuCode string) (*model.PaymentPackage, error) {
	var pkg model.PaymentPackage
	err := r.DB(ctx).Where("sku_code = ? AND deleted_at IS NULL", skuCode).First(&pkg).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &pkg, nil
}

func (r *paymentPackageRepository) ExistsBySKU(ctx context.Context, skuCode string) (bool, error) {
	var total int64
	err := r.DB(ctx).Model(&model.PaymentPackage{}).Where("sku_code = ?", skuCode).Count(&total).Error
	return total > 0, err
}

// ExistsByVirtualProductID 判断微信道具是否已被另一个未删除 SKU 占用。
// 同一道具只能对应一个本地 SKU，避免支付回调的商品、金额和权益出现歧义。
func (r *paymentPackageRepository) ExistsByVirtualProductID(ctx context.Context, virtualProductID string, excludeID int64) (bool, error) {
	var total int64
	db := r.DB(ctx).Model(&model.PaymentPackage{}).
		Where("virtual_product_id = ? AND deleted_at IS NULL", virtualProductID)
	if excludeID > 0 {
		db = db.Where("id != ?", excludeID)
	}
	err := db.Count(&total).Error
	return total > 0, err
}

func (r *paymentPackageRepository) CountByProductID(ctx context.Context, productID int64) (int64, error) {
	var total int64
	err := r.DB(ctx).Model(&model.PaymentPackage{}).
		Where("product_id = ? AND deleted_at IS NULL", productID).
		Count(&total).Error
	return total, err
}

func (r *paymentPackageRepository) List(ctx context.Context, query PaymentPackageListQuery) ([]*model.PaymentPackage, int64, error) {
	var (
		list  []*model.PaymentPackage
		total int64
	)
	db := r.DB(ctx).Model(&model.PaymentPackage{}).Where("deleted_at IS NULL")
	if query.ProductID != nil {
		db = db.Where("product_id = ?", *query.ProductID)
	}
	if query.Status != nil {
		db = db.Where("status = ?", *query.Status)
	}
	if query.Keyword != "" {
		like := "%" + query.Keyword + "%"
		db = db.Where("sku_code LIKE ? OR name LIKE ?", like, like)
	}
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if query.PageNum <= 0 {
		query.PageNum = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 20
	}
	offset := (query.PageNum - 1) * query.PageSize
	err := db.Order("sort DESC, create_at DESC").Offset(offset).Limit(query.PageSize).Find(&list).Error
	return list, total, err
}

func (r *paymentPackageRepository) ListAvailable(ctx context.Context, productID int64) ([]*model.PaymentPackage, error) {
	var list []*model.PaymentPackage
	err := r.DB(ctx).Model(&model.PaymentPackage{}).
		Where("deleted_at IS NULL AND product_id = ? AND status = ?", productID, model.PaymentPackageStatusPublished).
		Order("sort DESC, create_at ASC").
		Find(&list).Error
	return list, err
}

func (r *paymentPackageRepository) Update(ctx context.Context, pkg *model.PaymentPackage, expectedVersion int) (bool, error) {
	result := r.DB(ctx).Model(&model.PaymentPackage{}).
		Where("id = ? AND version = ? AND deleted_at IS NULL", pkg.ID, expectedVersion).
		Updates(map[string]any{
			"name":                 pkg.Name,
			"subtitle":             pkg.Subtitle,
			"badge":                pkg.Badge,
			"virtual_product_id":   pkg.VirtualProductID,
			"price_cents":          pkg.PriceCents,
			"original_price_cents": pkg.OriginalPriceCents,
			"benefit_config":       pkg.BenefitConfigJSON,
			"sale_rule":            pkg.SaleRuleJSON,
			"sort":                 pkg.Sort,
			"version":              gorm.Expr("version + 1"),
			"updated_by":           pkg.UpdatedBy,
			"update_at":            pkg.UpdateAt,
		})
	return result.RowsAffected == 1, result.Error
}

func (r *paymentPackageRepository) UpdateStatus(
	ctx context.Context,
	id int64,
	expectedVersion int,
	from, to model.PaymentPackageStatus,
	operator string,
) (bool, error) {
	result := r.DB(ctx).Model(&model.PaymentPackage{}).
		Where("id = ? AND version = ? AND status = ? AND deleted_at IS NULL", id, expectedVersion, from).
		Updates(map[string]any{
			"status":     to,
			"version":    gorm.Expr("version + 1"),
			"updated_by": operator,
			"update_at":  time.Now(),
		})
	return result.RowsAffected == 1, result.Error
}

func (r *paymentPackageRepository) SoftDelete(ctx context.Context, id int64, expectedVersion int, operator string) (bool, error) {
	now := time.Now()
	result := r.DB(ctx).Model(&model.PaymentPackage{}).
		Where("id = ? AND version = ? AND status != ? AND deleted_at IS NULL", id, expectedVersion, model.PaymentPackageStatusPublished).
		Updates(map[string]any{
			"deleted_at": now,
			"version":    gorm.Expr("version + 1"),
			"updated_by": operator,
			"update_at":  now,
		})
	return result.RowsAffected == 1, result.Error
}

func (r *paymentPackageRepository) CreateChangeLog(ctx context.Context, log *model.PaymentPackageChangeLog) error {
	return r.DB(ctx).Create(log).Error
}

func (r *paymentPackageRepository) ListChangeLogs(
	ctx context.Context,
	packageID int64,
	pageNum, pageSize int,
) ([]*model.PaymentPackageChangeLog, int64, error) {
	var (
		list  []*model.PaymentPackageChangeLog
		total int64
	)
	db := r.DB(ctx).Model(&model.PaymentPackageChangeLog{}).Where("sku_id = ?", packageID)
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if pageNum <= 0 {
		pageNum = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	err := db.Order("create_at DESC").Offset((pageNum - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}
