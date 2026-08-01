//go:build wireinject
// +build wireinject

package wire

import (
	"github.com/google/wire"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/go-nunu/nunu-layout-advanced/internal/repository"
	"github.com/go-nunu/nunu-layout-advanced/internal/server"
	"github.com/go-nunu/nunu-layout-advanced/internal/service"
	"github.com/go-nunu/nunu-layout-advanced/internal/task"
	"github.com/go-nunu/nunu-layout-advanced/pkg/app"
	"github.com/go-nunu/nunu-layout-advanced/pkg/log"
	"github.com/go-nunu/nunu-layout-advanced/pkg/sid"
)

var repositorySet = wire.NewSet(
	repository.NewDB,
	// repository.NewRedis,
	repository.NewRepository,
	repository.NewTransaction,
	repository.NewUserRepository,
	repository.NewJobRepository,
	repository.NewRentDetailRepository,
	repository.NewOrderRepository,
)

var taskSet = wire.NewSet(
	task.NewTask,
	task.NewUserTask,
	task.NewRentTask,
	service.NewPayService,
	provideZapLogger,
)
var serverSet = wire.NewSet(
	server.NewTaskServer,
)

// build App
func newApp(
	task *server.TaskServer,
) *app.App {
	return app.NewApp(
		app.WithServer(task),
		app.WithName("demo-task"),
	)
}

func provideZapLogger(logger *log.Logger) *zap.Logger {
	return logger.Logger
}

func NewWire(*viper.Viper, *log.Logger) (*app.App, func(), error) {
	panic(wire.Build(
		repositorySet,
		taskSet,
		serverSet,
		newApp,
		sid.NewSid,
	))
}
