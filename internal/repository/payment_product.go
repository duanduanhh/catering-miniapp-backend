package repository

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/go-nunu/nunu-layout-advanced/internal/model"
)

type PaymentProductWithCount struct {
	Product      *model.PaymentProduct
	PackageCount int64
}

type PaymentProductRepository interface {
	GetByID(ctx context.Context, id int64) (*model.PaymentProduct, error)
	GetByIDForUpdate(ctx context.Context, id int64) (*model.PaymentProduct, error)
	GetByCode(ctx context.Context, productCode string) (*model.PaymentProduct, error)
	List(ctx context.Context) ([]*model.PaymentProduct, error)
	ListWithPackageCount(ctx context.Context) ([]PaymentProductWithCount, error)
	UpdatePurchaseNotice(ctx context.Context, id int64, purchaseNotice string) error
}

func NewPaymentProductRepository(repository *Repository) PaymentProductRepository {
	return &paymentProductRepository{Repository: repository}
}

type paymentProductRepository struct {
	*Repository
}

func (r *paymentProductRepository) GetByID(ctx context.Context, id int64) (*model.PaymentProduct, error) {
	return r.get(ctx, r.DB(ctx).Where("id = ?", id))
}

func (r *paymentProductRepository) GetByIDForUpdate(ctx context.Context, id int64) (*model.PaymentProduct, error) {
	return r.get(ctx, r.DB(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", id))
}

func (r *paymentProductRepository) GetByCode(ctx context.Context, productCode string) (*model.PaymentProduct, error) {
	return r.get(ctx, r.DB(ctx).Where("product_code = ?", productCode))
}

func (r *paymentProductRepository) List(ctx context.Context) ([]*model.PaymentProduct, error) {
	var products []*model.PaymentProduct
	err := r.DB(ctx).Order("product_code ASC").Find(&products).Error
	return products, err
}

func (r *paymentProductRepository) ListWithPackageCount(ctx context.Context) ([]PaymentProductWithCount, error) {
	products, err := r.List(ctx)
	if err != nil {
		return nil, err
	}
	var counts []struct {
		ProductID int64
		Total     int64
	}
	if err := r.DB(ctx).Model(&model.PaymentPackage{}).
		Select("product_id, COUNT(*) AS total").
		Where("deleted_at IS NULL").
		Group("product_id").
		Scan(&counts).Error; err != nil {
		return nil, err
	}
	countsByProductID := make(map[int64]int64, len(counts))
	for _, count := range counts {
		countsByProductID[count.ProductID] = count.Total
	}
	result := make([]PaymentProductWithCount, 0, len(products))
	for _, product := range products {
		result = append(result, PaymentProductWithCount{
			Product:      product,
			PackageCount: countsByProductID[product.ID],
		})
	}
	return result, nil
}

func (r *paymentProductRepository) UpdatePurchaseNotice(ctx context.Context, id int64, purchaseNotice string) error {
	return r.DB(ctx).Model(&model.PaymentProduct{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"purchase_notice": purchaseNotice,
			"update_at":       time.Now(),
		}).Error
}

func (r *paymentProductRepository) get(ctx context.Context, db *gorm.DB) (*model.PaymentProduct, error) {
	var product model.PaymentProduct
	err := db.First(&product).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &product, nil
}
