package profiling

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	netpprof "net/http/pprof"
)

const (
	DefaultPort = 8095

	LabelComponent = "component"
)

type Config struct {
	Enable bool `env:"ENABLE" json:"enable" yaml:"enable"`
	Port   int  `env:"PORT" json:"port" yaml:"port"`
}

func RunHTTPServer(ctx context.Context, config Config, logger *slog.Logger) {

	if !config.Enable {
		return
	}

	logger = logger.With(LabelComponent, "profiling")

	if config.Port == 0 {
		config.Port = DefaultPort
	}

	go func() {
		mux := http.NewServeMux()

		// Add the pprof routes
		mux.HandleFunc("/debug/pprof/", netpprof.Index)
		mux.HandleFunc("/debug/pprof/cmdline", netpprof.Cmdline)
		mux.HandleFunc("/debug/pprof/profile", netpprof.Profile)
		mux.HandleFunc("/debug/pprof/symbol", netpprof.Symbol)
		mux.HandleFunc("/debug/pprof/trace", netpprof.Trace)

		mux.Handle("/debug/pprof/block", netpprof.Handler("block"))
		mux.Handle("/debug/pprof/goroutine", netpprof.Handler("goroutine"))
		mux.Handle("/debug/pprof/heap", netpprof.Handler("heap"))
		mux.Handle("/debug/pprof/threadcreate", netpprof.Handler("threadcreate"))

		srv := http.Server{
			Addr:        fmt.Sprintf(":%d", config.Port),
			Handler:     mux,
			ReadTimeout: 15 * time.Second,
		}

		go func() {
			if err := srv.ListenAndServe(); err != nil {
				logger.Error("profiling http server is down", "error", err)
				return
			}
		}()

		<-ctx.Done()

		ctxWithTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctxWithTimeout); err != nil {
			logger.Error("failed to close profiling http server", "error", err)
		}
	}()
}
