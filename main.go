package main

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/rs/zerolog"
)

const (
	defaultTickInterval = time.Second
	defaultPort         = "8000"
)

var commit = "dev"

func main() {
	ctx := context.Background()

	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	var tickInterval time.Duration

	tickIntervalStr := os.Getenv("TICK_INTERVAL")
	if tickIntervalStr == "" {
		tickInterval = defaultTickInterval
	} else {
		var err error
		tickInterval, err = time.ParseDuration(tickIntervalStr)
		if err != nil {
			logger.Fatal().Err(err).Msg("Failed to parse TICK_INTERVAL")
		}
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	logger.Info().
		Str("commit", commit).
		Dur("tickInterval", tickInterval).
		Str("port", port).
		Msg("Canary started")

	h := &netLogger{logger: &logger}

	go tick(ctx, &logger, tickInterval)

	if err := http.ListenAndServe(":"+port, h); err != nil {
		logger.Fatal().Err(err).Msg("Failed to start server")
	}

	logger.Info().Msg("Canary stopped")
}

func tick(ctx context.Context, logger *zerolog.Logger, interval time.Duration) {
	ticker := time.NewTicker(interval)

	for {
		select {
		case <-ticker.C:
			logger.Info().Msg("Tick")
		case <-ctx.Done():
			return
		}
	}
}

type netLogger struct {
	logger *zerolog.Logger
}

func (l *netLogger) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	l.logger.Info().Str("method", r.Method).Str("path", r.URL.Path).Msg("Request received")
	w.WriteHeader(http.StatusOK)
}
