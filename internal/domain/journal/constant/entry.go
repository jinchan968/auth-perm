package constant

// Period 日志时段
type Period string

const (
	PeriodMorning   Period = "晨"
	PeriodForenoon  Period = "上午"
	PeriodAfternoon Period = "下午"
	PeriodEvening   Period = "晚"
	PeriodNight     Period = "夜"
)

// Weather 天气
type Weather string

const (
	WeatherSunny  Weather = "晴"
	WeatherCloudy Weather = "多云"
	WeatherRainy  Weather = "雨"
	WeatherSnowy  Weather = "雪"
	WeatherFoggy  Weather = "雾"
	WeatherWindy  Weather = "风"
)

// InferPeriod 根据当前小时推断时段
func InferPeriod(hour int) Period {
	switch {
	case hour >= 6 && hour < 9:
		return PeriodMorning
	case hour >= 9 && hour < 12:
		return PeriodForenoon
	case hour >= 12 && hour < 18:
		return PeriodAfternoon
	case hour >= 18 && hour < 24:
		return PeriodEvening
	default:
		return PeriodNight
	}
}

// ValidPeriods 返回所有合法时段
func ValidPeriods() []Period {
	return []Period{PeriodMorning, PeriodForenoon, PeriodAfternoon, PeriodEvening, PeriodNight}
}

// IsValidPeriod 校验时段是否合法
func IsValidPeriod(p Period) bool {
	switch p {
	case PeriodMorning, PeriodForenoon, PeriodAfternoon, PeriodEvening, PeriodNight:
		return true
	default:
		return false
	}
}

// IsValidWeather 校验天气是否合法
func IsValidWeather(w Weather) bool {
	switch w {
	case WeatherSunny, WeatherCloudy, WeatherRainy, WeatherSnowy, WeatherFoggy, WeatherWindy:
		return true
	default:
		return false
	}
}
