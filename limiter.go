package golimiter

import (
	"context"
	"time"

	"golang.org/x/time/rate"
)

type Limiter struct {
	limiter *rate.Limiter
}

var defaultContext = context.Background()

// 限速单位为字节
func NewLimiter(limit int) *Limiter {
	return &Limiter{
		limiter: rate.NewLimiter(rate.Every(time.Second/time.Duration(limit)), limit), // 控制产生1个令牌的纳秒数，更均匀
	}
}

// 等待指定数量的令牌
func (l *Limiter) WaitToken(len int) error {
	maxLen := l.limiter.Burst()
	for {
		if len < maxLen {
			maxLen = len
		}

		if err := l.limiter.WaitN(defaultContext, maxLen); err != nil {
			return err
		}
		len -= maxLen
		if len == 0 {
			// 申请到足够量，返回
			return nil
		}
	}
}
