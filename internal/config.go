package internal

import (
	"github.com/spf13/viper"
)

type Config struct {
	Host        string `yaml:"host"`
	Port        string `yaml:"port"`
	Destination string `yaml:"destination"`
}

func (c Config) Validate() bool {
	return c.Host == "" || c.Port == ""
}

func InitConfig(path string) (Config, error) {
	viper.SetConfigFile(path)
	// viper.SetConfigType(cfgFileExt[1:])

	if err := viper.ReadInConfig(); err != nil {
		return Config{}, err
	} else {
		var cfg Config
		if err := viper.Unmarshal(&cfg); err != nil {
			return Config{}, err
		}
		return cfg, nil
	}
}
