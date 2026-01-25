package configs

type Config struct {
	Env          string             `mapstructure:"env" yaml:"env"`
	Integrations IntegrationsConfig `mapstructure:"integrations" yaml:"integrations"`
}

type IntegrationsConfig struct {
	Feishu FeishuConfig `mapstructure:"feishu" yaml:"feishu"`
}
