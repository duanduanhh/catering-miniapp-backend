package task

import (
	"context"
	"time"

	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/go-nunu/nunu-layout-advanced/internal/model"
	"github.com/go-nunu/nunu-layout-advanced/internal/repository"
)

// RentTask 招租相关定时任务。
type RentTask interface {
	// CleanupPending 清理超过 rent.pending_ttl_minutes 未支付的招租 job（软删除 + 硬删 rent_detail）。
	CleanupPending(ctx context.Context) error
}

func NewRentTask(
	task *Task,
	conf *viper.Viper,
	jobRepo repository.JobRepository,
	rentRepo repository.RentDetailRepository,
) RentTask {
	return &rentTask{
		Task:     task,
		conf:     conf,
		jobRepo:  jobRepo,
		rentRepo: rentRepo,
	}
}

type rentTask struct {
	*Task
	conf     *viper.Viper
	jobRepo  repository.JobRepository
	rentRepo repository.RentDetailRepository
}

func (t *rentTask) CleanupPending(ctx context.Context) error {
	ttlMin := t.conf.GetInt("rent.pending_ttl_minutes")
	if ttlMin <= 0 {
		ttlMin = 30
	}
	before := time.Now().Add(-time.Duration(ttlMin) * time.Minute)
	jobs, err := t.jobRepo.ListPendingRentBefore(ctx, before, 200)
	if err != nil {
		t.logger.Error("ListPendingRentBefore error", zap.Error(err))
		return err
	}
	if len(jobs) == 0 {
		return nil
	}
	cleaned := 0
	for _, j := range jobs {
		err := t.tm.Transaction(ctx, func(ctx context.Context) error {
			j.Status = model.JobStatusDeleted
			if err := t.jobRepo.Update(ctx, j); err != nil {
				return err
			}
			return t.rentRepo.DeleteByJobID(ctx, j.ID)
		})
		if err != nil {
			t.logger.Warn("cleanup pending rent failed", zap.Int64("job_id", j.ID), zap.Error(err))
			continue
		}
		cleaned++
	}
	t.logger.Info("CleanupPendingRent done", zap.Int("cleaned", cleaned), zap.Int("total", len(jobs)))
	return nil
}
