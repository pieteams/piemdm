package configs

type FeishuConfig struct {
	AppID           string `mapstructure:"app_id" yaml:"app_id"`
	AppSecret       string `mapstructure:"app_secret" yaml:"app_secret"`
	AdminInternalID string `mapstructure:"admin_internal_id" yaml:"admin_internal_id"` // 默认的发起人 ID (OpenID 或 UserID)
	Enabled         bool   `mapstructure:"enabled" yaml:"enabled"`
}
