package main

import (
	"log"
	"net/http"

	"sup-anapa/backend/internal/config"
	httpHandler "sup-anapa/backend/internal/http"
	"sup-anapa/backend/internal/repository"
	"sup-anapa/backend/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}
	repo := repository.New()
	weather := service.NewWeatherService(repo, cfg.WeatherAPIURL, cfg.WeatherCacheMin)
	h := httpHandler.NewHandler(repo, weather)
	mux := http.NewServeMux()
	h.Register(mux)
	log.Printf("backend started on :%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, mux); err != nil {
		log.Fatal(err)
	}
}
