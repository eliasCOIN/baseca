package config

import (
	"errors"
	"fmt"

	"github.com/coinbase/baseca/internal/logger"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type configProvider struct {
	v *viper.Viper
}

var _ ConfigProvider = (*configProvider)(nil)

func BuildViper(path string) (*viper.Viper, error) {
	ctxLogger := logger.ContextLogger{Logger: logger.DefaultLogger}
	ctxLogger.Info("Setting up Viper to load configuration", zap.String("config-path", path))

	v := viper.New()
	v.SetConfigFile(path)
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	return v, nil
}

func LoadConfig(viper *viper.Viper) (*Config, error) {
	if viper == nil {
		return nil, errors.New("Failed to load config.")
	}

	c := Config{}
	if err := viper.Unmarshal(&c); err != nil {
		return nil, errors.New("Failed to read configuration file.")
	}
	return &c, nil
}

func NewConfigProviderFromViper(v *viper.Viper) ConfigProvider {
	return &configProvider{v: v}
}

func (cp *configProvider) Get(path string, cfg interface{}) error {
	if !cp.Exists(path) {
		return fmt.Errorf("Path %s is not found in configuration.", path)
	}

	if err := cp.v.UnmarshalKey(path, cfg, func(setting *mapstructure.DecoderConfig) {
		setting.ErrorUnused = true
		setting.ZeroFields = true
	}); err != nil {
		return err
	}

	if u, ok := cfg.(interface{ Validate() error }); ok {
		if err := u.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (cp *configProvider) Exists(path string) bool {
	return cp.v.Get(path) != nil
}
