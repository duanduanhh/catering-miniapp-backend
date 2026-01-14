package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

func NewConfig(p string) *viper.Viper {
	envConf := os.Getenv("APP_CONF")
	if envConf == "" {
		appEnv := strings.ToLower(strings.TrimSpace(os.Getenv("APP_ENV")))
		if appEnv != "" {
			envConf = fmt.Sprintf("config/%s.yml", appEnv)
		} else {
			envConf = p
		}
	}
	fmt.Println("load conf file:", envConf)
	return getConfig(envConf)
}

func getConfig(path string) *viper.Viper {
	conf := viper.New()
	conf.SetConfigFile(path)
	err := conf.ReadInConfig()
	if err != nil {
		panic(err)
	}
	return conf
}
