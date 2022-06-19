package main

import (
	"cmd/main/app.go/internal/config"
	"cmd/main/app.go/internal/handlers/auth"
	"cmd/main/app.go/internal/handlers/metric"
	"cmd/main/app.go/pkg/cache/freecache"
	"cmd/main/app.go/pkg/logging"
	"cmd/main/app.go/pkg/shutdown"
	"errors"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"net"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"syscall"
	"time"
)

func main() {
	logging.Init()
	logger := logging.GetLogger()
	logger.Println("Инициализация логгера")

	logger.Println("Конфиг инит")
	cfg := config.GetConfig()
	logger.Println("Руты инит")
	router := httprouter.New()

	logger.Println("Кэш инит")
	refreshTokenCache := freecache.NewCacheRepo(104857600) // 100MB

	logger.Println("Создание хэндлеров")
	authHandler := auth.Handler{RTCache: refreshTokenCache, Logger: logger}
	authHandler.Register(router)

	metricHandler := metric.Handler{Logger: logger}
	metricHandler.Register(router)

	logger.Println("Приложение успешно запущено")
	start(router, logger, cfg)
}

func start(router *httprouter.Router, logger logging.Logger, cfg *config.Config) {
	var server *http.Server
	var listener net.Listener

	if cfg.Listen.Type == "sock" {
		appDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
		if err != nil {
			logger.Fatal(err)
		}
		socketPath := path.Join(appDir, "app.sock")
		logger.Infof("Сокет: %s", socketPath)

		logger.Info("Создание и прослушивание сокета")
		listener, err = net.Listen("unix", socketPath)
		if err != nil {
			logger.Fatal(err)
		}
	} else { //  Для локальной разработки используем порт
		logger.Infof("%s:%s", cfg.Listen.BindIP, cfg.Listen.Port)

		var err error

		listener, err = net.Listen("tcp", fmt.Sprintf("%s:%s", cfg.Listen.BindIP, cfg.Listen.Port))
		if err != nil {
			logger.Fatal(err)
		}
	}

	server = &http.Server{
		Handler:      router,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	go shutdown.Graceful([]os.Signal{syscall.SIGABRT, syscall.SIGQUIT, syscall.SIGHUP, os.Interrupt, syscall.SIGTERM},
		server)

	logger.Println("Приложение запущено")

	if err := server.Serve(listener); err != nil {
		switch {
		case errors.Is(err, http.ErrServerClosed):
			logger.Warn("Приложение остановлено")
		default:
			logger.Fatal(err)
		}
	}
}
