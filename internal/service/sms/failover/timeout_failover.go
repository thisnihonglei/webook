package failover

import (
	"context"
	"sync/atomic"
	"webook/internal/service/sms"
)

type TimeoutFailOverSMSService struct {
	svcs []sms.Service
	// 当前使用节点
	idx int32
	// 连续几个超时
	cnt int32
	// 切换的阈值，只读的
	threshold int32
}

func NewTimeoutFailOverSMSService(svcs []sms.Service, threshold int32) *TimeoutFailOverSMSService {
	return &TimeoutFailOverSMSService{
		svcs:      svcs,
		threshold: threshold,
	}
}

func (t *TimeoutFailOverSMSService) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	idx := atomic.LoadInt32(&t.idx)
	cnt := atomic.LoadInt32(&t.cnt)
	// 超过阈值，执行切换
	if cnt >= t.threshold {
		newIdx := (idx + 1) % int32(len(t.svcs))
		if atomic.CompareAndSwapInt32(&t.idx, idx, newIdx) {
			// 重置cnt这个计数
			atomic.StoreInt32(&t.cnt, 0)
		}
		idx = newIdx
	}
	svc := t.svcs[idx]
	err := svc.Send(ctx, tplId, args, numbers...)

	switch err {
	case nil:
		// 连续超时，所以不超时的时候重置为0
		atomic.StoreInt32(&t.cnt, 0)
		return nil
	case context.DeadlineExceeded:
		atomic.AddInt32(&t.cnt, 1)
	default:

	}
	return err
}
