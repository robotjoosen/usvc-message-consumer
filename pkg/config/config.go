// Package config applies environment vars to a struct.
package config

import (
	"errors"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

func Load(env interface{}, scope map[string]any) (interface{}, error) {
	viper.AddConfigPath(".")
	viper.SetConfigName(".env")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	for key, value := range scope {
		viper.SetDefault(key, value)
	}

	err := viper.ReadInConfig()
	if err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) {
			return env, err
		}
	}

	err = viper.Unmarshal(&env, func(config *mapstructure.DecoderConfig) {
		config.IgnoreUntaggedFields = true
	})

	return env, err
}
