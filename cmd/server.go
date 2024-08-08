package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fredytarigan/Tendo/pkg/tendo/config"
	"github.com/fredytarigan/Tendo/pkg/tendo/logger"
	"github.com/fredytarigan/Tendo/pkg/tendo/watcher"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

func ServerListen(kubeconfig string) {
	cfg := config.LoadConfig()

	ctx := context.Background()

	handler := mux.NewRouter();

	handler.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		logger.Logger.Info(
			"incoming request",
			zap.String("path", "/"),
		)
		fmt.Fprintf(w, "Hello World !!!")
	})

	initServeHttp(handler)
	errs := make(chan error)


	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errs <- fmt.Errorf("%s", <-c)
	}()

	go func() {
		address := fmt.Sprintf("%s:%s", cfg.AppHost, cfg.AppPort)

		logger.Logger.Info(fmt.Sprintf("Server is running and listening on %s", address))

		srvHttp := &http.Server{
			ReadTimeout: 5 * time.Second,
			WriteTimeout: 5 * time.Second,
			Addr: address,
			Handler: handler,
		}

		errs <- srvHttp.ListenAndServe()
	}()

	go func() {
		// wait for http server is running
		time.Sleep(5 * time.Second)
		logger.Logger.Info("Start running watcher service")

		watcher.Start(ctx, &cfg, kubeconfig)
	}()

	logger.Logger.Fatal(fmt.Sprintf("Received unrecovered errors, %s", <-errs))
}

func initServeHttp(handler *mux.Router) {
	handler.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Hello World !!!")
	})

	handler.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Server is ready and healthy")
	})
}