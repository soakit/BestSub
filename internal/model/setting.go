package model

const (
	// 系统设置
	SettingProxyEnable     = "proxy_enable"      // 全局代理开关
	SettingProxyURL        = "proxy_url"         // 全局代理地址
	SettingSubConvertUrl   = "sub_convert_url"   // 订阅转换地址
	SettingSubConvertProxy = "sub_convert_proxy" // 是否使用代理访问订阅转换
	SettingBindInterface   = "bind_interface"    // 网卡绑定
	// 测试设置
	SettingHealthCheckURL = "health_check_url" // 测活链接
	SettingSpeedTestURL   = "speed_test_url"   // 测速连接
	SettingMaxConcurrent  = "max_concurrent"   // 最大并发数
	// 前端设置
	SettingThemeAuto = "theme_auto" // 自动切换主题
	SettingTheme     = "theme"      // 主题 (light/dark)
)

type Setting struct {
	Key   string `gorm:"column:key;primaryKey;type:varchar(64)" json:"key"` // 配置键
	Value string `gorm:"column:value;type:text" json:"value"`               // 配置值
}
