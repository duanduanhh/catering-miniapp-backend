package server

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/go-nunu/nunu-layout-advanced/internal/model"
	"github.com/go-nunu/nunu-layout-advanced/pkg/log"
)

type MigrateServer struct {
	db  *gorm.DB
	log *log.Logger
}

func NewMigrateServer(db *gorm.DB, log *log.Logger) *MigrateServer {
	return &MigrateServer{db: db, log: log}
}

func (m *MigrateServer) Start(ctx context.Context) error {
	m.db.Exec("ALTER TABLE collect RENAME COLUMN `type` TO `biz_type`")
	if err := m.renamePaymentSchema(); err != nil {
		return err
	}
	if err := m.db.AutoMigrate(&model.PaymentProduct{}); err != nil {
		return err
	}
	if err := m.seedPaymentProducts(); err != nil {
		return err
	}
	if err := m.db.AutoMigrate(
		&model.User{}, &model.Enterprise{}, &model.Job{}, &model.Report{},
		&model.ContactFeedback{}, &model.CallbackHistory{}, &model.RentDetail{},
		&model.PaymentPackage{}, &model.PaymentPackageChangeLog{}, &model.OrderItem{},
	); err != nil {
		return err
	}
	if err := m.dropObsoletePaymentColumns(); err != nil {
		m.log.Error("obsolete payment columns cleanup error", zap.Error(err))
		return err
	}
	if err := m.validateSingleProductCardinality(); err != nil {
		return err
	}
	m.log.Info("AutoMigrate success")
	os.Exit(0)
	return nil
}

func (m *MigrateServer) validateSingleProductCardinality() error {
	var conflicts int64
	if err := m.db.Raw(`
		SELECT COUNT(*) FROM (
			SELECT payment_sku.product_id
			FROM payment_sku
			INNER JOIN payment_product ON payment_product.id = payment_sku.product_id
			WHERE payment_product.selection_mode = ? AND payment_sku.deleted_at IS NULL
			GROUP BY payment_sku.product_id HAVING COUNT(*) > 1
		) AS duplicated_single_products
	`, model.PaymentSelectionModeSingle).Scan(&conflicts).Error; err != nil {
		return err
	}
	if conflicts > 0 {
		return fmt.Errorf("%d single-selection products have more than one SKU", conflicts)
	}
	return nil
}

// dropObsoletePaymentColumns removes fields no longer part of the SKU model.
func (m *MigrateServer) dropObsoletePaymentColumns() error {
	if m.db.Migrator().HasColumn("payment_product", "business_type") {
		if err := m.db.Migrator().DropColumn(&model.PaymentProduct{}, "business_type"); err != nil {
			return err
		}
	}
	for _, column := range []string{"business_type", "description", "applicable_biz_types", "is_recommended", "effective_start_at", "effective_end_at"} {
		if m.db.Migrator().HasColumn("payment_sku", column) {
			if err := m.db.Migrator().DropColumn(&model.PaymentPackage{}, column); err != nil {
				return err
			}
		}
	}
	return nil
}

// renamePaymentSchema upgrades the old package terminology without creating a second data set.
func (m *MigrateServer) renamePaymentSchema() error {
	if m.db.Dialector.Name() != "mysql" {
		return nil
	}
	if m.db.Migrator().HasTable("payment_package") && !m.db.Migrator().HasTable("payment_sku") {
		if err := m.db.Exec("RENAME TABLE payment_package TO payment_sku").Error; err != nil {
			return err
		}
	}
	if m.db.Migrator().HasTable("payment_package_change_log") && !m.db.Migrator().HasTable("payment_sku_change_log") {
		if err := m.db.Exec("RENAME TABLE payment_package_change_log TO payment_sku_change_log").Error; err != nil {
			return err
		}
	}
	if m.db.Migrator().HasColumn("payment_sku_change_log", "package_id") {
		if err := m.db.Exec("ALTER TABLE payment_sku_change_log RENAME COLUMN package_id TO sku_id").Error; err != nil {
			return err
		}
	}
	if m.db.Migrator().HasColumn("payment_sku_change_log", "package_version") {
		if err := m.db.Exec("ALTER TABLE payment_sku_change_log RENAME COLUMN package_version TO sku_version").Error; err != nil {
			return err
		}
	}
	if m.db.Migrator().HasColumn("order_item", "package_id") {
		if err := m.db.Exec("ALTER TABLE order_item RENAME COLUMN package_id TO sku_id").Error; err != nil {
			return err
		}
	}
	if m.db.Migrator().HasColumn("order_item", "package_version") {
		if err := m.db.Exec("ALTER TABLE order_item RENAME COLUMN package_version TO sku_version").Error; err != nil {
			return err
		}
	}
	return nil
}

func (m *MigrateServer) seedPaymentProducts() error {
	now := time.Now()
	products := []*model.PaymentProduct{
		{ProductCode: model.PaymentProductCodeJobTop, Name: "岗位置顶", SelectionMode: model.PaymentSelectionModeMultiple, PurchaseNotice: "置顶权益支付成功后立即生效，具体时长以所选套餐为准。", CreateAt: now, UpdateAt: now},
		{ProductCode: model.PaymentProductCodeContactVoucher, Name: "联系券", SelectionMode: model.PaymentSelectionModeMultiple, PurchaseNotice: "联系券绑定当前账号，不可转让；支付成功后自动到账。", CreateAt: now, UpdateAt: now},
		{ProductCode: model.PaymentProductCodePaidRefresh, Name: "付费刷新", SelectionMode: model.PaymentSelectionModeSingle, PurchaseNotice: "付费刷新支付成功后立即提升当前信息的排序时间。", CreateAt: now, UpdateAt: now},
		{ProductCode: model.PaymentProductCodeRentPublish, Name: "招租发布", SelectionMode: model.PaymentSelectionModeSingle, PurchaseNotice: "支付成功后招租信息自动发布。", CreateAt: now, UpdateAt: now},
	}
	if err := m.db.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "product_code"}}, DoNothing: true}).Create(&products).Error; err != nil {
		return err
	}
	for _, product := range products {
		if err := m.db.Model(&model.PaymentProduct{}).
			Where("product_code = ? AND (purchase_notice = '' OR purchase_notice IS NULL)", product.ProductCode).
			Update("purchase_notice", product.PurchaseNotice).Error; err != nil {
			return err
		}
	}
	return nil
}

func (m *MigrateServer) Stop(ctx context.Context) error {
	m.log.Info("AutoMigrate stop")
	return nil
}
