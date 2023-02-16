package limit

import (
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"
)

/*
假设n秒内最多处理b个请求。那么会最多缓存 b 个通过的请求与对应的时间戳，假设这个缓存集合为B。
每当有请求到来时，从B中删除掉n秒前的所有请求，查看集合是否满了，如果没满，则通过请求，并放入集合，如果满了就触发拒绝策略。

记录下所有的请求时间点，系统会将这些日志存储在按照时间先后来排序的集合中。
新请求到来时，会先判断指定时间范围内的请求数量是否超过阈值，超出阈值的请求会被丢弃。
这种方式避免了固定窗口算法容易遇到的请求突变问题，限流比较准确。

不过因为它要记录下每次请求的时间点，所以会额外消耗内存与 CPU。

下面的代码给出了滑动日志算法的示例，在该示例中，我们可以指定多个限速策略，
例如策略 A 在 1 分钟内允许处理 100 个请求，并且策略 B 在 1 小时内允许处理 1000 个请求。
*/

// ViolationStrategyError 违背策略错误
type ViolationStrategyError struct {
	Limit  int           // 窗口请求上限
	Window time.Duration // 窗口时间大小
}

func (e *ViolationStrategyError) Error() string {
	return fmt.Sprintf("violation strategy that limit = %d and window = %d", e.Limit, e.Window)
}

// SlidingLogLimiterStrategy 滑动日志限流器的策略
type SlidingLogLimiterStrategy struct {
	limit        int   // 窗口请求上限
	window       int64 // 窗口时间大小
	smallWindows int64 // 小窗口数量
}

func NewSlidingLogLimiterStrategy(limit int, window time.Duration) *SlidingLogLimiterStrategy {
	return &SlidingLogLimiterStrategy{
		limit:  limit,
		window: int64(window),
	}
}

// SlidingLogLimiter 滑动日志限流器
type SlidingLogLimiter struct {
	strategies  []*SlidingLogLimiterStrategy // 滑动日志限流器策略列表
	smallWindow int64                        // 小窗口时间大小
	counters    map[int64]int                // 小窗口计数器
	mutex       sync.Mutex                   // 避免并发问题
}

func NewSlidingLogLimiter(smallWindow time.Duration, strategies ...*SlidingLogLimiterStrategy) (*SlidingLogLimiter, error) {
	// 复制策略避免被修改
	strategies = append(make([]*SlidingLogLimiterStrategy, 0, len(strategies)), strategies...)

	// 不能不设置策略
	if len(strategies) == 0 {
		return nil, errors.New("must be set strategies")
	}

	// 排序策略，窗口时间大的排前面，相同窗口上限大的排前面
	sort.Slice(strategies, func(i, j int) bool {
		a, b := strategies[i], strategies[j]
		if a.window == b.window {
			return a.limit > b.limit
		}
		return a.window > b.window
	})
	fmt.Println(strategies[0], strategies[1])

	for i, strategy := range strategies {
		// 随着窗口时间变小，窗口上限也应该变小
		if i > 0 {
			if strategy.limit >= strategies[i-1].limit {
				return nil, errors.New("the smaller window should be the smaller limit")
			}
		}
		// 窗口时间必须能够被小窗口时间整除
		if strategy.window%int64(smallWindow) != 0 {
			return nil, errors.New("window cannot be split by integers")
		}
		strategy.smallWindows = strategy.window / int64(smallWindow)
	}

	return &SlidingLogLimiter{
		strategies:  strategies,
		smallWindow: int64(smallWindow),
		counters:    make(map[int64]int),
	}, nil
}

func (l *SlidingLogLimiter) TryAcquire() error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	// 获取当前小窗口值
	currentSmallWindow := time.Now().UnixNano() / l.smallWindow * l.smallWindow
	// 获取每个策略的起始小窗口值
	startSmallWindows := make([]int64, len(l.strategies))
	for i, strategy := range l.strategies {
		startSmallWindows[i] = currentSmallWindow - l.smallWindow*(strategy.smallWindows-1)
	}

	// 计算每个策略当前窗口的请求总数
	counts := make([]int, len(l.strategies))
	for smallWindow, counter := range l.counters {
		if smallWindow < startSmallWindows[0] {
			delete(l.counters, smallWindow)
			continue
		}
		for i := range l.strategies {
			if smallWindow >= startSmallWindows[i] {
				counts[i] += counter
			}
		}
	}

	// 若到达对应策略窗口请求上限，请求失败，返回违背的策略
	for i, strategy := range l.strategies {
		if counts[i] >= strategy.limit {
			return &ViolationStrategyError{
				Limit:  strategy.limit,
				Window: time.Duration(strategy.window),
			}
		}
	}

	// 若没到窗口请求上限，当前小窗口计数器+1，请求成功
	l.counters[currentSmallWindow]++
	return nil
}
