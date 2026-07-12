package server

import (
	"context"
	"time"

	"github.com/go-co-op/gocron"
	"go.uber.org/zap"

	"github.com/go-nunu/nunu-layout-advanced/internal/task"
	"github.com/go-nunu/nunu-layout-advanced/pkg/log"
)

type TaskServer struct {
	log       *log.Logger
	scheduler *gocron.Scheduler
	userTask  task.UserTask
	rentTask  task.RentTask
}

func NewTaskServer(
	log *log.Logger,
	userTask task.UserTask,
	rentTask task.RentTask,
) *TaskServer {
	return &TaskServer{
		log:      log,
		userTask: userTask,
		rentTask: rentTask,
	}
}

func (t *TaskServer) Start(ctx context.Context) error {
	gocron.SetPanicHandler(func(jobName string, recoverData any) {
		t.log.Error("TaskServer Panic", zap.String("job", jobName), zap.Any("recover", recoverData))
	})

	// eg: crontab task
	t.scheduler = gocron.NewScheduler(time.UTC)
	// if you are in China, you will need to change the time zone as follows
	// t.scheduler = gocron.NewScheduler(time.FixedZone("PRC", 8*60*60))

	// _, err := t.scheduler.Every("3s").Do(func()
	_, err := t.scheduler.CronWithSeconds("0/3 * * * * *").Do(func() {
		err := t.userTask.CheckUser(ctx)
		if err != nil {
			t.log.Error("CheckUser error", zap.Error(err))
		}
	})
	if err != nil {
		t.log.Error("CheckUser error", zap.Error(err))
	}

	// 招租超时清理：每 5 分钟扫描一次，清理超过 rent.pending_ttl_minutes 未支付的招租 job。
	_, err = t.scheduler.Cron("*/5 * * * *").Do(func() {
		if err := t.rentTask.CleanupPending(ctx); err != nil {
			t.log.Error("CleanupPending rent error", zap.Error(err))
		}
	})
	if err != nil {
		t.log.Error("register CleanupPending rent error", zap.Error(err))
	}

	t.scheduler.StartBlocking()
	return nil
}

func (t *TaskServer) Stop(ctx context.Context) error {
	t.scheduler.Stop()
	t.log.Info("TaskServer stop...")
	return nil
}
