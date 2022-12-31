package limiter

import (
	"context"
	"sort"
	"time"

	"golang.org/x/time/rate"
)

type RateLimiter interface {
	// Wait 等待可用的令牌
	Wait(ctx context.Context) error
	// Limit 速率
	Limit() rate.Limit
}

// Per 每隔多长时间生成多少个token
func Per(eventCount int, duration time.Duration) rate.Limit {
	return rate.Every(duration / time.Duration(eventCount))
}

// MultiLimiter 多层限流器
type MultiLimiter struct {
	limiters []RateLimiter
}

// Wait 循环遍历多层限速器 MultiLimiter 中所有的限速器并索要令牌，
// 只有当所有的限速器规则都满足后，才会正常执行后续的操作。
func (l *MultiLimiter) Wait(ctx context.Context) error {
	for _, l := range l.limiters {
		if err := l.Wait(ctx); err != nil {
			return err
		}
	}

	return nil
}

func (l *MultiLimiter) Limit() rate.Limit {
	return l.limiters[0].Limit()
}

// Multi 函数用于聚合多个 RateLimiter，并将速率由小到大排序
func Multi(limiters ...RateLimiter) *MultiLimiter {
	byLimit := func(i, j int) bool {
		return limiters[i].Limit() < limiters[j].Limit()
	}

	sort.Slice(limiters, byLimit)

	return &MultiLimiter{limiters: limiters}
}
