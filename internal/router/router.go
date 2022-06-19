package router

import (
	"cmd/main/app.go/internal/app_context"
	"cmd/main/app.go/internal/auth"
	"cmd/main/app.go/pkg/logging"
	"cmd/main/app.go/pkg/metric"
	"cmd/main/app.go/pkg/middleware/jwt"
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

func Init() {
	logger := logging.GetLogger()
	logger.Println("Инициализация логгера...")

	router := httprouter.New()

	router.HandlerFunc("POST", auth.URL, auth.Auth)

	router.HandlerFunc("GET", metric.HEARTBEAT_URL, jwt.JWTMiddleware(metric.Heartbeat))
	router.HandlerFunc("GET", metric.TEST_URL, metric.Test)

	ctx := app_context.GetInstance()
	cfg := ctx.Config

	var server *http.Server
	var listener net.Listener

	if cfg.Listen.Type == "sock" {
		appDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
		if err != nil {
			logger.Fatal(err)
		}
		socketPath := path.Join(appDir, "app.sock")
		logger.Infof("socket path: %s", socketPath)

		logger.Info("create and listen unix socket")
		listener, err = net.Listen("unix", socketPath)
		if err != nil {
			logger.Fatal(err)
		}
	} else {
		logger.Infof("хост: %s порт: %s", cfg.Listen.BindIP, cfg.Listen.Port)

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

	if err := server.Serve(listener); err != nil {
		switch {
		case errors.Is(err, http.ErrServerClosed):
			logger.Warn("остановка сервера")
		default:
			logger.Fatal(err)
		}
	}
}
