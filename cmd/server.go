package cmd

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fredytarigan/Tendo/pkg/tendo/config"
	"github.com/fredytarigan/Tendo/pkg/tendo/watcher"
	"github.com/gorilla/mux"
)

func ServerListen(kubeconfig string) {
	cfg := config.LoadConfig()

	ctx := context.Background()

	handler := mux.NewRouter();

	handler.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
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

		fmt.Printf("Server is running and listening on %s \n", address)

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
		fmt.Println("Start running watcher service")

		watcher.Start(ctx, &cfg, kubeconfig)
	}()

	log.Fatalf("Received unrecovered errors, %s", <-errs)
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