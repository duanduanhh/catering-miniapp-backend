package router

import (
	"github.com/go-nunu/nunu-layout-advanced/internal/handler"
	"github.com/go-nunu/nunu-layout-advanced/pkg/jwt"
	"github.com/go-nunu/nunu-layout-advanced/pkg/log"
	"github.com/spf13/viper"
)

type RouterDeps struct {
	Logger               *log.Logger
	Config               *viper.Viper
	JWT                  *jwt.JWT
	UserHandler          *handler.UserHandler
	JobHandler           *handler.JobHandler
	CollectHandler       *handler.CollectHandler
	ContactHistoryHandler *handler.ContactHistoryHandler
	CostHistoryHandler   *handler.CostHistoryHandler
	WechatHandler        *handler.WechatHandler
	UploadHandler        *handler.UploadHandler
}
