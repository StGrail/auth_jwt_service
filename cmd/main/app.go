package main

import (
	"cmd/main/app.go/internal/router"
	"cmd/main/app.go/pkg/logging"
)

func main() {
	logging.Init()
	logger := logging.GetLogger()
	logger.Println("Инициализация логгера...")

	defer router.Init()

	logger.Println("Приложение успешно запущено")
}
