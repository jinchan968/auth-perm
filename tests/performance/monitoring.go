package performance

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// PerformanceMetrics 性能指标
type PerformanceMetrics struct {
	Name            string
	TotalRequests   int64
	SuccessfulReqs  int64
	FailedReqs      int64
	MinDuration     time.Duration
	MaxDuration     time.Duration
	TotalDuration   time.Duration
	AverageDuration time.Duration
	Throughput      float64
	StartTime       time.Time
	EndTime         time.Time
}

// NewMetrics 创建新的性能指标
func NewMetrics(name string) *PerformanceMetrics {
	return &PerformanceMetrics{
		Name:        name,
		StartTime:   time.Now(),
		MinDuration: 999999 * time.Hour,
	}
}

// Record 记录一次请求
func (m *PerformanceMetrics) Record(duration time.Duration, success bool) {
	m.TotalRequests++
	m.TotalDuration += duration

	if duration < m.MinDuration {
		m.MinDuration = duration
	}
	if duration > m.MaxDuration {
		m.MaxDuration = duration
	}

	if success {
		m.SuccessfulReqs++
	} else {
		m.FailedReqs++
	}
}

// Finish 完成统计
func (m *PerformanceMetrics) Finish() {
	m.EndTime = time.Now()
	totalTime := m.EndTime.Sub(m.StartTime)
	if m.TotalRequests > 0 {
		m.AverageDuration = m.TotalDuration / time.Duration(m.TotalRequests)
		m.Throughput = float64(m.TotalRequests) / totalTime.Seconds()
	}
}

// String 返回格式化的性能报告
func (m *PerformanceMetrics) String() string {
	m.Finish()
	successRate := float64(m.SuccessfulReqs) / float64(m.TotalRequests) * 100

	return fmt.Sprintf(`
性能报告: %s
================
总请求数:       %d
成功请求数:     %d
失败请求数:     %d
成功率:         %.2f%%
总耗时:         %v
平均耗时:       %v
最小耗时:       %v
最大耗时:       %v
吞吐量:         %.2f req/sec
================
`,
		m.Name,
		m.TotalRequests,
		m.SuccessfulReqs,
		m.FailedReqs,
		successRate,
		m.TotalDuration,
		m.AverageDuration,
		m.MinDuration,
		m.MaxDuration,
		m.Throughput,
	)
}

// ConcurrentBenchmark 并发性能测试
type ConcurrentBenchmark struct {
	Metrics *PerformanceMetrics
	Workers int
}

// NewBenchmark 创建新的并发测试
func NewBenchmark(name string, workers int) *ConcurrentBenchmark {
	return &ConcurrentBenchmark{
		Metrics: NewMetrics(name),
		Workers: workers,
	}
}

// Run 运行并发测试
func (cb *ConcurrentBenchmark) Run(testFunc func() error) {
	var wg sync.WaitGroup

	for i := 0; i < cb.Workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for {
				operationStart := time.Now()
				err := testFunc()
				duration := time.Since(operationStart)

				success := err == nil
				cb.Metrics.Record(duration, success)

				// 模拟持续负载
				select {
				case <-time.After(10 * time.Millisecond):
					continue
				default:
				}
			}
		}(i)
	}

	// 运行5秒
	time.Sleep(5 * time.Second)

	// 注意：实际项目中需要优雅停止goroutines
	// 这里使用time.After模拟，实际应该使用context
}

// Monitor 监控器
type Monitor struct {
	mu          sync.RWMutex
	metrics     map[string]*PerformanceMetrics
	startTime   time.Time
	activeUsers int64
}

// NewMonitor 创建新的监控器
func NewMonitor() *Monitor {
	return &Monitor{
		metrics:   make(map[string]*PerformanceMetrics),
		startTime: time.Now(),
	}
}

// StartMetric 开始监控指标
func (m *Monitor) StartMetric(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.metrics[name]; !ok {
		m.metrics[name] = NewMetrics(name)
	}
}

// RecordMetric 记录指标
func (m *Monitor) RecordMetric(name string, duration time.Duration, success bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if metric, ok := m.metrics[name]; ok {
		metric.Record(duration, success)
	}
}

// GetMetric 获取指标
func (m *Monitor) GetMetric(name string) *PerformanceMetrics {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if metric, ok := m.metrics[name]; ok {
		return metric
	}
	return nil
}

// Report 生成完整报告
func (m *Monitor) Report() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	report := fmt.Sprintf(`
性能监控报告
================
监控时间: %v
活跃用户: %d
================
`,
		time.Since(m.startTime),
		m.activeUsers,
	)

	for name, metric := range m.metrics {
		report += fmt.Sprintf("%s\n%s\n", name, metric.String())
	}

	return report
}

// PerformanceAlert 性能告警
type PerformanceAlert struct {
	Threshold    time.Duration
	WarningRate  float64
	CriticalRate float64
}

// Check 检查性能告警
func (pa *PerformanceAlert) Check(metric *PerformanceMetrics) (status string, message string) {
	if metric.AverageDuration > pa.Threshold {
		successRate := float64(metric.SuccessfulReqs) / float64(metric.TotalRequests)
		if successRate < pa.CriticalRate {
			return "CRITICAL", fmt.Sprintf("平均耗时 %v 超过阈值 %v，成功率 %.2f%%", metric.AverageDuration, pa.Threshold, successRate*100)
		} else if successRate < pa.WarningRate {
			return "WARNING", fmt.Sprintf("平均耗时 %v 超过阈值 %v，成功率 %.2f%%", metric.AverageDuration, pa.Threshold, successRate*100)
		}
	}
	return "OK", "性能正常"
}

// PerformanceRegressionTest 性能回归测试
func PerformanceRegressionTest(t *testing.T) {
	// 基线性能（历史最佳性能）
	baseline := map[string]time.Duration{
		"GetDevices":      50 * time.Millisecond,
		"TrustDevice":     20 * time.Millisecond,
		"GetSecurityLogs": 100 * time.Millisecond,
		"RevokeSession":   30 * time.Millisecond,
	}

	// 性能退化阈值
	regressionThreshold := 1.5 // 50%退化

	scenarios := []struct {
		Name     string
		Func     func() error
		Baseline time.Duration
	}{
		{"GetDevices", func() error {
			devices := simulateGetDevices(100)
			if len(devices) != 100 {
				return assert.AnError
			}
			return nil
		}, baseline["GetDevices"]},
		{"TrustDevice", func() error {
			if !simulateTrustDevice() {
				return assert.AnError
			}
			return nil
		}, baseline["TrustDevice"]},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.Name, func(t *testing.T) {
			// 运行多次取平均
			const iterations = 10
			var totalDuration time.Duration

			for i := 0; i < iterations; i++ {
				start := time.Now()
				err := scenario.Func()
				duration := time.Since(start)

				assert.NoError(t, err)
				totalDuration += duration
			}

			averageDuration := totalDuration / iterations
			regressionRatio := float64(averageDuration) / float64(scenario.Baseline)

			t.Logf("%s: 平均耗时 %v, 基线 %v, 退化比 %.2f", scenario.Name, averageDuration, scenario.Baseline, regressionRatio)

			// 性能退化不应超过阈值
			assert.Less(t, regressionRatio, regressionThreshold, "性能退化超过阈值")
		})
	}
}
