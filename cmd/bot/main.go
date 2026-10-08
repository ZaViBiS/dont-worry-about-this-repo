package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"
	"time"
	_ "time/tzdata"

	"ZaViBiS/dont-worry-about-this-repo/internal/bot"
	"ZaViBiS/dont-worry-about-this-repo/internal/config"
	database "ZaViBiS/dont-worry-about-this-repo/internal/db"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}).With().Caller().Logger()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("error loading envs")
	}

	db, err := database.DBInit(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("db init")
	}
	defer db.Close()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Info().Msg("starting bot")
	if err := bot.Run(ctx, db, cfg); err != nil && !errors.Is(err, context.Canceled) {
		log.Fatal().Err(err).Msg("bot failed")
	}

	log.Info().Msg("shutdown")
}
