package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/bir/iken/config"
	"github.com/bir/iken/errs"
	"github.com/bir/iken/httplog"
	"github.com/go-chi/chi/v5"
	"github.com/lavaai/kit/auth"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	chiTrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/go-chi/chi.v5"

	"{{ .Params.Package }}"
)

type Config struct {
	Debug               bool          `env:"DEBUG"`
	Port                int           `env:"PORT, 3500"`
	HttpWriteTimeout    time.Duration `env:"HTTP_WRITE_TIMEOUT, 30s"`
	HttpReadTimeout     time.Duration `env:"HTTP_READ_TIMEOUT, 30s"`
	HttpIdleTimeout     time.Duration `env:"HTTP_IDLE_TIMEOUT, 50s"`
	HttpShutdownTimeout time.Duration `env:"HTTP_SHUTDOWN_TIMEOUT, 5s"`
}

func main() {
	var cfg Config
	err := config.Load(&cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("loading config")
	}

	l := setupLogging(true)

	router := chi.NewRouter().With(
		httplog.RecoverLogger(l),
		chiTrace.Middleware(),
		httplog.RequestLogger(httplog.LogAll),
	)

	svc :=  {{ $.PackageName }}.New()
	{{ $.PackageName }}.RegisterHTTP(svc, router
{{- if .HasAuthentication -}}
	{{- range $security, $value := .File.API.Components.SecuritySchemes -}}
	, {{ $.PackageName }}.{{ pascal $security }}Auth({{ pascal $security }}Auth)
	{{- end -}}
    {{- if .HasAuthorization -}}
	, Authorize
    {{- end -}}
{{- end -}}
)

	httpServer := http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		WriteTimeout: cfg.HttpWriteTimeout,
		ReadTimeout:  cfg.HttpReadTimeout,
		IdleTimeout:  cfg.HttpIdleTimeout,
		Handler:      router,
	}

	httpServerExit := make(chan int, 1)

	go func() {
		defer func() { httpServerExit <- 1 }()

		l.Info().Msgf("Serving on: http://%s", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				log.Error().Stack().Err(err).Msg("HTTP Server error")
			}
		}
		log.Info().Msg("HTTP Server stopped")
	}()

	sigInt := make(chan os.Signal, 1)

	signal.Notify(sigInt, os.Interrupt) // We'll start graceful shutdowns when quit via SIGINT (Ctrl+C)

	var wg sync.WaitGroup // Block until we receive any signal.

	select {
	case <-sigInt:
		shutdownServer(&httpServer, cfg.HttpShutdownTimeout, &wg)
		log.Info().Msg("SIGINT received, shutting down.")
	case <-httpServerExit:
		log.Info().Msg("HTTP Server exited")
	}

	wg.Wait()

}

func setupLogging(consoleLog bool) zerolog.Logger {
	zerolog.DurationFieldInteger = true
	zerolog.DurationFieldUnit = time.Millisecond
	zerolog.ErrorStackMarshaler = errs.MarshalStack
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	if consoleLog {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	zerolog.DefaultContextLogger = &log.Logger

	return log.Logger
}

func shutdownServer(server *http.Server, duration time.Duration, wg *sync.WaitGroup) {
	wg.Add(1)

	go func() {
		defer wg.Done()

		ctx, cancel := context.WithTimeout(context.Background(), duration)
		defer cancel()

		err := server.Shutdown(ctx)
		if err != nil {
			log.Error().Stack().Err(err).Msg("Error shutting down server.")
		}
	}()
}

{{- if .HasAuthentication -}}
	{{- range $security, $value := .File.API.Components.SecuritySchemes }}

func {{ pascal $security }}Auth(ctx context.Context, key string) (*{{ $.CheckPackage $.Params.Auth "" }}, error){
	return nil, {{ $.PackageName }}.ErrNotImplemented
}
	{{- end -}}
    {{- if .HasAuthentication }}

func Authorize(ctx context.Context, user *auth.LavaUser, scopes []string) error {
	return form.ErrNotImplemented
}

{{- end -}}
{{- end -}}