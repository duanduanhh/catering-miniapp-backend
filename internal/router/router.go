package router

import (
	"github.com/spf13/viper"

	"github.com/go-nunu/nunu-layout-advanced/internal/handler"
	"github.com/go-nunu/nunu-layout-advanced/internal/repository"
	"github.com/go-nunu/nunu-layout-advanced/pkg/jwt"
	"github.com/go-nunu/nunu-layout-advanced/pkg/log"
)

type RouterDeps struct {
	Logger                       *log.Logger
	Config                       *viper.Viper
	JWT                          *jwt.JWT
	UserRepo                     repository.UserRepository
	UserHandler                  *handler.UserHandler
	JobHandler                   *handler.JobHandler
	OrderHandler                 *handler.OrderHandler
	CollectHandler               *handler.CollectHandler
	ContactHistoryHandler        *handler.ContactHistoryHandler
	ContactVoucherHistoryHandler *handler.ContactVoucherHistoryHandler
	WechatHandler                *handler.WechatHandler
	UploadHandler                *handler.UploadHandler
	PositionCategoryHandler      *handler.PositionCategoryHandler
	FeedbackHandler              *handler.FeedbackHandler
	EnterpriseHandler            *handler.EnterpriseHandler
	AdminJobHandler              *handler.AdminJobHandler
	AdminAuthHandler             *handler.AdminAuthHandler
	AdminListHandler             *handler.AdminListHandler
	ReportHandler                *handler.ReportHandler
	ContactFeedbackHandler       *handler.ContactFeedbackHandler
	PaymentPackageHandler        *handler.PaymentPackageHandler
}
