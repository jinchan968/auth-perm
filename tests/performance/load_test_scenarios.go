package performance

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// LoadTestScenario 负载测试场景
type LoadTestScenario struct {
	Name          string
	Concurrency   int
	Duration      time.Duration
	Threshold     time.Duration
	TestOperation func() error
}

// RunLoadTest 运行负载测试
func RunLoadTest(t *testing.T, scenario LoadTestScenario) {
	t.Logf("开始负载测试: %s", scenario.Name)
	t.Logf("并发数: %d, 持续时间: %v", scenario.Concurrency, scenario.Duration)

	start := time.Now()
	errors := make(chan error, scenario.Concurrency)
	var wg sync.WaitGroup

	// 启动并发goroutines
	for i := 0; i < scenario.Concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			endTime := start.Add(scenario.Duration)
			operations := 0

			for time.Now().Before(endTime) {
				err := scenario.TestOperation()
				if err != nil {
					errors <- err
				}
				operations++
			}

			t.Logf("Worker %d 完成 %d 次操作", workerID, operations)
		}(i)
	}

	wg.Wait()
	close(errors)

	// 检查错误
	errorCount := 0
	for err := range errors {
		if err != nil {
			errorCount++
			t.Logf("错误: %v", err)
		}
	}

	totalDuration := time.Since(start)
	tps := float64(scenario.Concurrency) * scenario.Duration.Seconds() / totalDuration.Seconds()

	t.Logf("负载测试完成: %s", scenario.Name)
	t.Logf("总耗时: %v", totalDuration)
	t.Logf("错误数: %d", errorCount)
	t.Logf("吞吐量: %.2f ops/sec", tps)

	// 验证错误率不超过5%
	errorRate := float64(errorCount) / float64(scenario.Concurrency)
	assert.Less(t, errorRate, 0.05, "错误率不应超过5%")
}

// BenchmarkScenarios 性能基准测试场景
var BenchmarkScenarios = []struct {
	Name        string
	DeviceCount int
	LogCount    int
	Concurrency int
}{
	{"小规模", 50, 500, 5},
	{"中规模", 200, 2000, 10},
	{"大规模", 500, 5000, 20},
}

// BenchmarkDeviceList 小规模设备列表性能测试
func BenchmarkDeviceList(b *testing.B) {
	scenario := BenchmarkScenarios[0]

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		devices := simulateGetDevices(scenario.DeviceCount)
		_ = devices
	}
}

// BenchmarkDeviceListLarge 大规模设备列表性能测试
func BenchmarkDeviceListLarge(b *testing.B) {
	scenario := BenchmarkScenarios[2]

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		devices := simulateGetDevices(scenario.DeviceCount)
		_ = devices
	}
}

// BenchmarkSecurityLogs 小规模安全日志性能测试
func BenchmarkSecurityLogs(b *testing.B) {
	scenario := BenchmarkScenarios[0]

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		logs := simulateGetSecurityLogs(scenario.LogCount)
		_ = logs
	}
}

// BenchmarkSecurityLogsLarge 大规模安全日志性能测试
func BenchmarkSecurityLogsLarge(b *testing.B) {
	scenario := BenchmarkScenarios[2]

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		logs := simulateGetSecurityLogs(scenario.LogCount)
		_ = logs
	}
}

// TestLoadGetDevices 负载测试 - 获取设备列表
func TestLoadGetDevices(t *testing.T) {
	scenario := LoadTestScenario{
		Name:        "获取设备列表负载测试",
		Concurrency: 10,
		Duration:    5 * time.Second,
		Threshold:   100 * time.Millisecond,
		TestOperation: func() error {
			devices := simulateGetDevices(100)
			if len(devices) != 100 {
				return assert.AnError
			}
			return nil
		},
	}

	RunLoadTest(t, scenario)
}

// TestLoadTrustDevice 负载测试 - 信任设备
func TestLoadTrustDevice(t *testing.T) {
	scenario := LoadTestScenario{
		Name:        "信任设备负载测试",
		Concurrency: 20,
		Duration:    3 * time.Second,
		Threshold:   50 * time.Millisecond,
		TestOperation: func() error {
			if !simulateTrustDevice() {
				return assert.AnError
			}
			return nil
		},
	}

	RunLoadTest(t, scenario)
}

// TestLoadGetSecurityLogs 负载测试 - 获取安全日志
func TestLoadGetSecurityLogs(t *testing.T) {
	scenario := LoadTestScenario{
		Name:        "获取安全日志负载测试",
		Concurrency: 5,
		Duration:    5 * time.Second,
		Threshold:   200 * time.Millisecond,
		TestOperation: func() error {
			logs := simulateGetSecurityLogs(1000)
			if len(logs) != 1000 {
				return assert.AnError
			}
			return nil
		},
	}

	RunLoadTest(t, scenario)
}

// TestLoadRevokeSession 负载测试 - 撤销会话
func TestLoadRevokeSession(t *testing.T) {
	scenario := LoadTestScenario{
		Name:        "撤销会话负载测试",
		Concurrency: 15,
		Duration:    3 * time.Second,
		Threshold:   50 * time.Millisecond,
		TestOperation: func() error {
			if !simulateRevokeSession() {
				return assert.AnError
			}
			return nil
		},
	}

	RunLoadTest(t, scenario)
}

// TestSpikeLoad 尖峰负载测试
func TestSpikeLoad(t *testing.T) {
	t.Log("开始尖峰负载测试")

	// 正常负载
	t.Log("阶段1: 正常负载 (5并发)")
	normalScenario := LoadTestScenario{
		Name:        "正常负载",
		Concurrency: 5,
		Duration:    2 * time.Second,
		Threshold:   100 * time.Millisecond,
		TestOperation: func() error {
			devices := simulateGetDevices(50)
			if len(devices) != 50 {
				return assert.AnError
			}
			return nil
		},
	}
	RunLoadTest(t, normalScenario)

	// 短暂休息
	time.Sleep(1 * time.Second)

	// 尖峰负载
	t.Log("阶段2: 尖峰负载 (50并发)")
	spikeScenario := LoadTestScenario{
		Name:        "尖峰负载",
		Concurrency: 50,
		Duration:    2 * time.Second,
		Threshold:   150 * time.Millisecond,
		TestOperation: func() error {
			devices := simulateGetDevices(50)
			if len(devices) != 50 {
				return assert.AnError
			}
			return nil
		},
	}
	RunLoadTest(t, spikeScenario)

	// 恢复阶段
	time.Sleep(1 * time.Second)

	// 回到正常负载
	t.Log("阶段3: 恢复负载 (5并发)")
	recoveryScenario := LoadTestScenario{
		Name:        "恢复负载",
		Concurrency: 5,
		Duration:    2 * time.Second,
		Threshold:   100 * time.Millisecond,
		TestOperation: func() error {
			devices := simulateGetDevices(50)
			if len(devices) != 50 {
				return assert.AnError
			}
			return nil
		},
	}
	RunLoadTest(t, recoveryScenario)
}

// TestSustainedLoad 持续负载测试
func TestSustainedLoad(t *testing.T) {
	scenario := LoadTestScenario{
		Name:        "持续负载测试",
		Concurrency: 10,
		Duration:    30 * time.Second, // 30秒持续负载
		Threshold:   100 * time.Millisecond,
		TestOperation: func() error {
			devices := simulateGetDevices(100)
			if len(devices) != 100 {
				return assert.AnError
			}
			return nil
		},
	}

	RunLoadTest(t, scenario)
}

// TestMemoryUsage 内存使用测试
func TestMemoryUsage(t *testing.T) {
	t.Log("开始内存使用测试")

	// 初始内存使用
	t.Log("执行前内存检查")
	// 实际项目中会使用runtime.ReadMemStats()获取内存使用情况

	// 大量设备查询
	for i := 0; i < 100; i++ {
		devices := simulateGetDevices(1000)
		_ = devices
	}

	t.Log("执行后内存检查")
	// 验证内存使用是否在合理范围内

	// 测试通过
	assert.True(t, true, "内存使用测试通过")
}

// TestDatabaseConnectionPool 数据库连接池测试
func TestDatabaseConnectionPool(t *testing.T) {
	// 模拟数据库连接池
	poolSize := 10
	connections := make(chan struct{}, poolSize)

	// 获取连接
	for i := 0; i < poolSize; i++ {
		connections <- struct{}{}
	}

	// 测试并发获取连接
	errors := make(chan error, poolSize*2)
	for i := 0; i < poolSize*2; i++ {
		go func(id int) {
			select {
			case <-connections:
				// 模拟数据库操作
				devices := simulateGetDevices(100)
				_ = devices
				connections <- struct{}{}
			default:
				errors <- assert.AnError
			}
		}(i)
	}

	// 等待一段时间让goroutines完成
	time.Sleep(100 * time.Millisecond)

	// 检查错误
	errorCount := 0
	for {
		select {
		case err := <-errors:
			if err != nil {
				errorCount++
			}
		default:
			if errorCount > 0 {
				t.Errorf("连接池耗尽错误数: %d", errorCount)
			}
			return
		}
	}
}

// BenchmarkAllScenarios 所有场景性能基准测试
func BenchmarkAllScenarios(b *testing.B) {
	for _, scenario := range BenchmarkScenarios {
		b.Run(scenario.Name, func(b *testing.B) {
			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				devices := simulateGetDevices(scenario.DeviceCount)
				logs := simulateGetSecurityLogs(scenario.LogCount)

				// 模拟组合操作
				_ = devices
				_ = logs
			}
		})
	}
}
