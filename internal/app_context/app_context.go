package app_context

import (
	"cmd/main/app.go/internal/config"
	"cmd/main/app.go/pkg/logging"
	"sync"
)

type AppContext struct {
	Config *config.Config
}

var instance *AppContext
var once sync.Once

func GetInstance() *AppContext {
	once.Do(func() {
		logging.GetLogger().Println("Инициализация контекста приложения...")
		instance = &AppContext{
			Config: config.GetConfig(),
		}
	})

	return instance
}
