package main

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"sup-anapa/backend/internal/config"
	"sup-anapa/backend/internal/db"
	httpHandler "sup-anapa/backend/internal/http"
	"sup-anapa/backend/internal/repository"
	"sup-anapa/backend/internal/service"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})

	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load config")
	}

	database, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect db")
	}

	repo := repository.New(database)
	weather := service.NewWeatherService(repo, cfg.WeatherAPIURL, cfg.WeatherCacheMin)

	r := gin.Default()
	h := httpHandler.NewHandler(repo, weather)
	h.Register(r)

	log.Info().Str("port", cfg.Port).Msg("backend started")
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal().Err(err).Msg("server failed")
	}
}
