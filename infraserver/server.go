package infraserver

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	DefaultPort              = 8090
	DefaultMetricsEndpoint   = "/metrics"
	DefaultReadinessEndpoint = "/ready"
	DefaultLivenessEndpoint  = "/alive"
)

type (
	Server struct {
		Config
		server *http.Server
	}

	Config struct {
		Port              int           `env:"PORT"  json:"port" json:"port" yaml:"port"`
		MetricsEndpoint   string        `env:"METRICS_ENDPOINT" json:"metrics_endpoint" yaml:"metrics_endpoint"`
		ReadinessEndpoint string        `env:"READINESS_ENDPOINT" json:"readiness_endpoint" yaml:"readiness_endpoint"`
		LivenessEndpoint  string        `env:"LIVENESS_ENDPOINT" json:"liveness_endpoint" yaml:"liveness_endpoint"`
		Timeout           time.Duration `env:"TIMEOUT" envDefault:"5s" yaml:"timeout" default:"5s" `
		CheckTimeout      time.Duration `env:"CHECK_TIMEOUT" envDefault:"30s" json:"check_timeout" yaml:"check_timeout" default:"30s" `
	}
)

func New(config Config) *Server {

	if config.Port == 0 {
		config.Port = DefaultPort
	}

	if config.MetricsEndpoint == "" {
		config.MetricsEndpoint = DefaultMetricsEndpoint
	}

	if config.ReadinessEndpoint == "" {
		config.ReadinessEndpoint = DefaultReadinessEndpoint
	}

	if config.LivenessEndpoint == "" {
		config.LivenessEndpoint = DefaultLivenessEndpoint
	}

	return &Server{
		Config: config,
	}
}

func (s *Server) Run(ctx context.Context, readiness []func(context.Context) error, liveness []func(context.Context) error) error {
	mux := http.NewServeMux()
	mux.Handle(s.MetricsEndpoint, promhttp.Handler())
	mux.Handle(s.ReadinessEndpoint, s.createHandler(readiness))
	mux.Handle(s.LivenessEndpoint, s.createHandler(liveness))

	s.server = &http.Server{
		Addr:         fmt.Sprintf(":%d", s.Port),
		Handler:      mux,
		ReadTimeout:  s.Timeout,
		WriteTimeout: s.Timeout,
	}

	chanErr := make(chan error)
	go func() {
		defer close(chanErr)

		if err := s.server.ListenAndServe(); err != nil {
			chanErr <- err
		}
	}()

	var resErr error

	select {
	case <-ctx.Done():
	case err := <-chanErr:
		if !errors.Is(err, context.Canceled) {
			resErr = err
		}
	}

	_ = s.close()
	return resErr
}

func (s *Server) close() error {
	return s.server.Close()
}

func (s *Server) createHandler(handlers []func(context.Context) error) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(context.Background(), s.CheckTimeout)
		defer cancel()

		var resErrs []error
		for _, h := range handlers {
			if err := h(ctx); err != nil {
				resErrs = append(resErrs, err)
			}
		}

		err := errors.Join(resErrs...)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(fmt.Sprintf("healthcheck error: %v", err)))

			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
}
