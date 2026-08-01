//go:build wireinject
// +build wireinject

package wire

import (
	"github.com/google/wire"
	"github.com/spf13/viper"

	"github.com/go-nunu/nunu-layout-advanced/internal/handler"
	"github.com/go-nunu/nunu-layout-advanced/internal/job"
	"github.com/go-nunu/nunu-layout-advanced/internal/repository"
	"github.com/go-nunu/nunu-layout-advanced/internal/router"
	"github.com/go-nunu/nunu-layout-advanced/internal/server"
	"github.com/go-nunu/nunu-layout-advanced/internal/service"
	"github.com/go-nunu/nunu-layout-advanced/pkg/app"
	"github.com/go-nunu/nunu-layout-advanced/pkg/jwt"
	"github.com/go-nunu/nunu-layout-advanced/pkg/log"
	"github.com/go-nunu/nunu-layout-advanced/pkg/server/http"
	"github.com/go-nunu/nunu-layout-advanced/pkg/sid"
	"github.com/go-nunu/nunu-layout-advanced/pkg/wechatpay"
)

var repositorySet = wire.NewSet(
	repository.NewDB,
	// repository.NewRedis,
	// repository.NewMongo,
	repository.NewRepository,
	repository.NewTransaction,
	repository.NewUserRepository,
	repository.NewJobRepository,
	repository.NewCollectRepository,
	repository.NewContactHistoryRepository,
	repository.NewOrderRepository,
	repository.NewOrderItemRepository,
	repository.NewContactVoucherHistoryRepository,
	repository.NewPositionCategoryRepository,
	repository.NewFeedbackRepository,
	repository.NewEnterpriseRepository,
	repository.NewReportRepository,
	repository.NewContactFeedbackRepository,
	repository.NewCallbackHistoryRepository,
	repository.NewRentDetailRepository,
	repository.NewPaymentProductRepository,
	repository.NewPaymentPackageRepository,
)

var serviceSet = wire.NewSet(
	service.NewService,
	service.NewUserService,
	service.NewJobService,
	service.NewCollectService,
	service.NewContactHistoryService,
	service.NewOrderService,
	service.NewOrderItemService,
	service.NewContactVoucherHistoryService,
	service.NewWechatService,
	service.NewUploadService,
	service.NewImageSecurityService,
	NewPayServiceProvider,
	service.NewPositionCategoryService,
	service.NewFeedbackService,
	service.NewEnterpriseService,
	service.NewAdminJobService,
	service.NewAdminAuthService,
	service.NewAdminListService,
	service.NewReportService,
	service.NewContactFeedbackService,
	service.NewCallbackHistoryService,
	service.NewPaymentPackageService,
)

var handlerSet = wire.NewSet(
	handler.NewHandler,
	handler.NewUserHandler,
	handler.NewJobHandler,
	handler.NewOrderHandler,
	handler.NewCollectHandler,
	handler.NewContactHistoryHandler,
	handler.NewContactVoucherHistoryHandler,
	handler.NewWechatHandler,
	handler.NewUploadHandler,
	handler.NewPositionCategoryHandler,
	handler.NewFeedbackHandler,
	handler.NewEnterpriseHandler,
	handler.NewAdminJobHandler,
	handler.NewAdminAuthHandler,
	handler.NewAdminListHandler,
	handler.NewReportHandler,
	handler.NewContactFeedbackHandler,
	handler.NewPaymentPackageHandler,
)

var wechatPaySet = wire.NewSet(
	NewWechatPayClientProvider,
)

var jobSet = wire.NewSet(
	job.NewJob,
	job.NewUserJob,
)
var serverSet = wire.NewSet(
	server.NewHTTPServer,
	server.NewJobServer,
)

// NewPayServiceProvider 提供 PayService
func NewPayServiceProvider(config *viper.Viper, logger *log.Logger) (service.PayService, error) {
	return service.NewPayService(config, logger.Logger)
}

// NewWechatPayClientProvider 提供 WechatPayClient
func NewWechatPayClientProvider(config *viper.Viper) (*wechatpay.WechatPayClient, error) {
	return wechatpay.NewWechatPayClient(config)
}

// build App
func newApp(
	httpServer *http.Server,
	jobServer *server.JobServer,
	// task *server.Task,
) *app.App {
	return app.NewApp(
		app.WithServer(httpServer, jobServer),
		app.WithName("demo-server"),
	)
}

func NewWire(*viper.Viper, *log.Logger) (*app.App, func(), error) {
	panic(wire.Build(
		repositorySet,
		serviceSet,
		handlerSet,
		wechatPaySet,
		jobSet,
		serverSet,
		wire.Struct(new(router.RouterDeps), "*"),
		sid.NewSid,
		jwt.NewJwt,
		newApp,
	))
}
