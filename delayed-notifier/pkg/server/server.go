package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

type HTTPServer struct {
	router http.Handler
}

func NewHTTPServer(router http.Handler) *HTTPServer {
	return &HTTPServer{router: router}
}

func (s *HTTPServer) GracefulRun(ctx context.Context,host string, port int) error {
	httpServer := &http.Server{
		Addr:    fmt.Sprintf("%s:%d",host, port),
		Handler: s.router,
	}
	serverStopped := make(chan bool, 1)
	signalListenerExited := make(chan bool, 1)
	go listenSignal(ctx, httpServer, serverStopped, signalListenerExited)

	err := httpServer.ListenAndServe()
	serverStopped <- true
	<-signalListenerExited
	if err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("error while listening HTTP port '%d': %w", port, err)
		}
	}
	return nil

}

func listenSignal(ctx context.Context, httpServer *http.Server, serverStopped <-chan bool, funcExited chan<- bool) {
	signalCtx, stop := signal.NotifyContext(ctx, os.Interrupt)
	defer stop()
	select {
	case <-serverStopped:
		break
	case <-signalCtx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err := httpServer.Shutdown(shutdownCtx)
		if err != nil {
			log.Fatalf("error while shutting down http server: %v", err)
		}

	}
	funcExited <- true

}
