package configs

import (
	"github.com/spf13/viper"
)

type EnvConfig struct {
	AcessKey  string `mapstructure:"ACESS_KEY"`
	SecretKey string `mapstructure:"SECRET_KEY"`
	Bucket    string `mapstructure:"3Bucket"`
}

func LoadConfig(path string) (*EnvConfig, error) {
	var cfg EnvConfig

	viper.SetConfigType("env")
	viper.AddConfigPath(path)
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	err := viper.ReadInConfig()

	if err != nil {
		panic(err)
	}

	err = viper.Unmarshal(&cfg)

	if err != nil {
		panic(err)
	}

	return &cfg, nil
}
