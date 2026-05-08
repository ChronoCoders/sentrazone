package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ChronoCoders/sentra/internal/agent"
	"github.com/ChronoCoders/sentra/internal/alert"
	"github.com/ChronoCoders/sentra/internal/api"
	"github.com/ChronoCoders/sentra/internal/config"
	"github.com/ChronoCoders/sentra/internal/control"
	"github.com/ChronoCoders/sentra/internal/store"
	"github.com/ChronoCoders/sentra/internal/wireguard"
	"github.com/ChronoCoders/sentra/internal/ws"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})

	cfg := config.Load()

	db, err := store.New(cfg.DBPath, cfg.AdminEmail, cfg.AdminPassword)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to init database")
	}
	defer db.Close()

	wg, err := wireguard.NewWGManager(cfg.WGInterface)
	if err != nil {
		log.Error().Err(err).Msg("failed to init wireguard manager - local agent reporting will be limited")
	} else {
		defer wg.Close()
		if _, err := wg.GetStatus(context.Background()); err != nil {
			log.Error().Err(err).Msg("failed to get status from wireguard interface - local agent reporting will be limited")
		}
	}

	alerter := alert.New(cfg)

	bus := control.NewEventBus()

	hub := ws.NewHub()
	go hub.Run()

	client := control.NewStatusCache(bus, hub, alerter)

	histRecorder := control.NewHistoryRecorder(bus, db)
	go histRecorder.Run(context.Background())

	scheduler := control.NewScheduler(db, client, alerter)
	go scheduler.Run(context.Background())

	if !cfg.DisableAgent {
		reporter := agent.NewEventBusReporter(bus)
		var ag *agent.Agent
		if wg != nil {
			ag = agent.New(wg, reporter, "local")
		} else {
			ag = agent.New(nil, reporter, "local")
		}
		go func() {
			if err := ag.Run(context.Background()); err != nil {
				log.Error().Err(err).Msg("agent run error")
			}
		}()
	} else {
		log.Info().Msg("internal agent disabled by configuration")
	}

	srv := api.NewServer(cfg, db, client, hub, bus)

	httpServer := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: srv,
	}

	go func() {
		log.Info().Str("port", cfg.Port).Msg("starting control server")
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("server error")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info().Msg("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("server forced to shutdown")
	}

	log.Info().Msg("server exited")
}
