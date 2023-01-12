package app

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gorilla/mux"

	"github.com/kaz-as/test-get-items/config"
	"github.com/kaz-as/test-get-items/internal/getitems"
	"github.com/kaz-as/test-get-items/internal/middlewares"
	"github.com/kaz-as/test-get-items/pkg/httpserver"
)

func Run(cfg *config.Config) (ret error) {
	r := mux.NewRouter()
	r.StrictSlash(true)

	r.Use(
		mux.MiddlewareFunc(middlewares.Recoverer()),
		mux.MiddlewareFunc(middlewares.Logger()),
	)

	getItemsHandler, err := getitems.NewHandler(cfg.File)
	if err != nil {
		return fmt.Errorf("create handler for %s failed: %w", cfg.File, err)
	}

	r.Handle("/get-items", getItemsHandler).Methods(http.MethodGet)

	srv := httpserver.New(r,
		httpserver.Port(cfg.Port),
	)

	srv.Start()
	defer func(srv *httpserver.Server) {
		err := srv.Shutdown()
		if err != nil {
			log.Printf("srv shutdown: %s", err)
		}
	}(srv)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	select {
	case s := <-interrupt:
		log.Println("app - Run - signal: " + s.String())
	case err := <-srv.Notify():
		ret = fmt.Errorf("srv.Notify: %w", err)
		log.Println(ret)
	}

	return
}
