package performance

// simulateGetDevices 模拟获取设备列表
func simulateGetDevices(count int) []map[string]interface{} {
	devices := make([]map[string]interface{}, 0, count)
	for i := 0; i < count; i++ {
		device := map[string]interface{}{
			"device_id":   "device-fingerprint-" + string(rune(i)),
			"platform":    "Mac OS",
			"browser":     "Chrome",
			"trusted":     i%2 == 0,
			"last_active": "2026-01-06T15:30:00Z",
		}
		devices = append(devices, device)
	}
	return devices
}

// simulateTrustDevice 模拟信任设备
func simulateTrustDevice() bool {
	// 模拟数据库更新操作
	return true
}

// simulateGetSecurityLogs 模拟获取安全日志
func simulateGetSecurityLogs(count int) []map[string]interface{} {
	logs := make([]map[string]interface{}, 0, count)
	for i := 0; i < count; i++ {
		log := map[string]interface{}{
			"id":         "log-" + string(rune(i)),
			"action":     "login",
			"success":    true,
			"created_at": "2026-01-06T15:30:00Z",
		}
		logs = append(logs, log)
	}
	return logs
}

// simulateRevokeSession 模拟撤销会话
func simulateRevokeSession() bool {
	// 模拟数据库更新操作
	return true
}
