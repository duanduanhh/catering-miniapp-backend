package task

import (
	"context"
	"errors"
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
	orderRepo repository.OrderRepository,
) RentTask {
	return &rentTask{
		Task:      task,
		conf:      conf,
		jobRepo:   jobRepo,
		rentRepo:  rentRepo,
		orderRepo: orderRepo,
	}
}

type rentTask struct {
	*Task
	conf      *viper.Viper
	jobRepo   repository.JobRepository
	rentRepo  repository.RentDetailRepository
	orderRepo repository.OrderRepository
}

var errPendingRentOrderChanged = errors.New("pending rent order state changed")

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
		orderNos, err := t.orderRepo.ListPendingRentOrderNos(ctx, j.ID)
		if err != nil {
			t.logger.Warn("list pending rent orders failed", zap.Int64("job_id", j.ID), zap.Error(err))
			continue
		}
		// 订单已支付但岗位尚未激活时，不能依据“没有待支付订单”清理，保留给回调/人工补偿处理。
		if len(orderNos) == 0 {
			t.logger.Warn("skip pending rent cleanup because no pending order was found", zap.Int64("job_id", j.ID))
			continue
		}
		err = t.tm.Transaction(ctx, func(ctx context.Context) error {
			for _, orderNo := range orderNos {
				canceled, err := t.orderRepo.CancelPendingOrder(ctx, orderNo, "招租支付超时已关闭")
				if err != nil {
					return err
				}
				// 支付推送先到时，订单状态已被原子更新为已支付；本次清理必须整体回滚。
				if !canceled {
					return errPendingRentOrderChanged
				}
			}
			j.Status = model.JobStatusDeleted
			if err := t.jobRepo.Update(ctx, j); err != nil {
				return err
			}
			return t.rentRepo.DeleteByJobID(ctx, j.ID)
		})
		if err != nil {
			if errors.Is(err, errPendingRentOrderChanged) {
				continue
			}
			t.logger.Warn("cleanup pending rent failed", zap.Int64("job_id", j.ID), zap.Error(err))
			continue
		}
		cleaned++
	}
	t.logger.Info("CleanupPendingRent done", zap.Int("cleaned", cleaned), zap.Int("total", len(jobs)))
	return nil
}
