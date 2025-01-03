package configs

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/spf13/viper"
)

type configuration struct {
	Server struct {
		Host string
		Port string
	}
}

var Config *configuration

func NewConfiguration() (*configuration, error) {
	viper.AddConfigPath("settings")
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(`.`, `_`))
	err := viper.ReadInConfig()
	if err != nil {
		return nil, fmt.Errorf("error loading config file: %s", err)
	}
	err = viper.Unmarshal(&Config)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %s", err)
	}
	slog.Info(fmt.Sprintf("Servers in 'config.yaml' were added on service: %v", Config))
	return Config, nil
}
