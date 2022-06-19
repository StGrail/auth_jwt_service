package config

import (
	"cmd/main/app.go/pkg/logging"
	"github.com/ilyakaznacheev/cleanenv"
	"sync"
)

// Config Базовый конфиг, парсим config.yaml, логгируем в случае ошибки.
type Config struct {
	IsDebug *bool `yaml:"is_debug"`
	JWT     struct {
		Secret string `yaml:"secret" env-required:"true"`
	}
	Listen struct {
		Type   string `yaml:"type" env-default:"port"`
		BindIP string `yaml:"bind_ip" env-default:"127.0.0.1"`
		Port   string `yaml:"port" env-default:"5000"`
	}
}

var instance *Config
var once sync.Once

func GetConfig() *Config {
	once.Do(func() {
		logger := logging.GetLogger()
		logger.Info("Чтение конфига приложения")
		instance = &Config{}
		if err := cleanenv.ReadConfig("config.yaml", instance); err != nil {
			help, _ := cleanenv.GetDescription(instance, nil)
			logger.Info(help)
			logger.Fatal(err)
		}
	})
	return instance
}
