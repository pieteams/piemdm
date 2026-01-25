package configloader

import (
	"flag"
	"os"
	"strings"

	"github.com/spf13/viper"
)

func Load() (*viper.Viper, error) {
	envConf := os.Getenv("APP_CONF")
	if envConf == "" {
		// 检查是否在测试环境下运行
		if isTest() {
			// 在测试环境下，只有当前目录存在 config/local.yml 时才使用它
			if _, err := os.Stat("config/local.yml"); err == nil {
				envConf = "config/local.yml"
			}
		} else {
			if flag.Lookup("conf") == nil {
				flag.StringVar(&envConf, "conf", "config/local.yml", "config path, eg: -conf config/local.yml")
			}
			if !flag.Parsed() {
				flag.Parse()
			}
		}
	}
	// 非测试环境下默认使用 local 配置
	if envConf == "" && !isTest() {
		envConf = "local"
	}
	return getConfig(envConf)
}

func isTest() bool {
	return strings.HasSuffix(os.Args[0], ".test") ||
		strings.HasSuffix(os.Args[0], ".test.exe") ||
		strings.Contains(os.Args[0], "/_go_build_")
}

func getConfig(path string) (*viper.Viper, error) {
	conf := viper.New()
	if path != "" {
		conf.SetConfigFile(path)
	}
	conf.AutomaticEnv()
	conf.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if path != "" {
		err := conf.ReadInConfig()
		if err != nil {
			return nil, err
		}
	}
	return conf, nil
}
