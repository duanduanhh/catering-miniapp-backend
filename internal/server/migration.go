package server

import (
	"context"
	"os"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/go-nunu/nunu-layout-advanced/internal/model"
	"github.com/go-nunu/nunu-layout-advanced/pkg/log"
)

type MigrateServer struct {
	db  *gorm.DB
	log *log.Logger
}

func NewMigrateServer(db *gorm.DB, log *log.Logger) *MigrateServer {
	return &MigrateServer{
		db:  db,
		log: log,
	}
}

func (m *MigrateServer) Start(ctx context.Context) error {
	// rename collect.type → collect.biz_type (idempotent: ignore error if already renamed)
	m.db.Exec("ALTER TABLE collect RENAME COLUMN `type` TO `biz_type`")

	if err := m.db.AutoMigrate(
		&model.User{},
		&model.Enterprise{},
		&model.Job{},
		&model.Report{},
		&model.ContactFeedback{},
		&model.CallbackHistory{},
	); err != nil {
		m.log.Error("user migrate error", zap.Error(err))
		return err
	}
	m.log.Info("AutoMigrate success")
	os.Exit(0)
	return nil
}

func (m *MigrateServer) Stop(ctx context.Context) error {
	m.log.Info("AutoMigrate stop")
	return nil
}
