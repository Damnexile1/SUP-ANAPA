package service

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"time"

	"sup-anapa/backend/internal/models"
	"sup-anapa/backend/internal/repository"
)

type WeatherService struct {
	repo     *repository.Repository
	apiURL   string
	http     *http.Client
	cacheTTL time.Duration
}

type WeatherResponse struct {
	Temperature     float64           `json:"temperature"`
	WindSpeed       float64           `json:"wind_speed"`
	Precipitation   float64           `json:"precipitation"`
	CloudCover      int               `json:"cloud_cover"`
	ConditionsLevel string            `json:"conditions_level"`
	Explanation     string            `json:"explanation"`
	Score           int               `json:"score"`
	SuggestedSlots  []models.TimeSlot `json:"suggested_slots,omitempty"`
	Raw             map[string]any    `json:"raw,omitempty"`
}

func NewWeatherService(repo *repository.Repository, apiURL string, cacheTTL time.Duration) *WeatherService {
	return &WeatherService{repo: repo, apiURL: apiURL, cacheTTL: cacheTTL, http: &http.Client{Timeout: 10 * time.Second}}
}

func (s *WeatherService) Get(lat, lng float64, target time.Time, routeID, instructorID string) (WeatherResponse, error) {
	targetHour := target.UTC().Truncate(time.Hour)
	if cached, err := s.repo.FindWeatherSnapshot(lat, lng, targetHour, s.cacheTTL); err == nil {
		resp := mapSnapshot(cached)
		if resp.ConditionsLevel == "Плохие" {
			resp.SuggestedSlots, _ = s.repo.SuggestedSlots(target, routeID, instructorID, 5)
		}
		return resp, nil
	}
	apiData, err := s.fetch(lat, lng, targetHour)
	if err != nil {
		return WeatherResponse{}, err
	}
	score, level, explanation := scoreWeather(apiData.Temperature, apiData.WindSpeed, apiData.Precipitation, apiData.CloudCover)
	snapshot := models.WeatherSnapshot{LocationLat: lat, LocationLng: lng, TimeFrom: targetHour, TimeTo: targetHour.Add(time.Hour), Temperature: apiData.Temperature, WindSpeed: apiData.WindSpeed, Precipitation: apiData.Precipitation, CloudCover: apiData.CloudCover, ConditionsLevel: level, Score: score, Raw: apiData.Raw, FetchedAt: time.Now().UTC()}
	_ = s.repo.SaveWeatherSnapshot(&snapshot)
	resp := WeatherResponse{Temperature: apiData.Temperature, WindSpeed: apiData.WindSpeed, Precipitation: apiData.Precipitation, CloudCover: apiData.CloudCover, ConditionsLevel: level, Explanation: explanation, Score: score, Raw: apiData.Raw}
	if level == "Плохие" {
		resp.SuggestedSlots, _ = s.repo.SuggestedSlots(target, routeID, instructorID, 5)
	}
	return resp, nil
}

type fetchedData struct {
	Temperature, WindSpeed, Precipitation float64
	CloudCover                            int
	Raw                                   map[string]any
}

func (s *WeatherService) fetch(lat, lng float64, targetHour time.Time) (fetchedData, error) {
	u, _ := url.Parse(s.apiURL)
	q := u.Query()
	q.Set("latitude", fmt.Sprintf("%.5f", lat))
	q.Set("longitude", fmt.Sprintf("%.5f", lng))
	q.Set("hourly", "temperature_2m,wind_speed_10m,precipitation,cloud_cover")
	q.Set("timezone", "UTC")
	q.Set("start_date", targetHour.Format("2006-01-02"))
	q.Set("end_date", targetHour.Format("2006-01-02"))
	u.RawQuery = q.Encode()
	resp, err := s.http.Get(u.String())
	if err != nil {
		return fetchedData{}, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return fetchedData{}, fmt.Errorf("weather api error: %s", string(body))
	}
	var payload struct {
		Hourly struct {
			Time        []string  `json:"time"`
			Temperature []float64 `json:"temperature_2m"`
			Wind        []float64 `json:"wind_speed_10m"`
			Precip      []float64 `json:"precipitation"`
			Cloud       []int     `json:"cloud_cover"`
		} `json:"hourly"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return fetchedData{}, err
	}
	if len(payload.Hourly.Time) == 0 {
		return fetchedData{}, fmt.Errorf("empty weather payload")
	}
	idx := 0
	min := math.MaxFloat64
	for i, t := range payload.Hourly.Time {
		parsed, _ := time.Parse("2006-01-02T15:04", t)
		d := math.Abs(parsed.Sub(targetHour).Hours())
		if d < min {
			min = d
			idx = i
		}
	}
	raw := map[string]any{}
	_ = json.Unmarshal(body, &raw)
	return fetchedData{Temperature: payload.Hourly.Temperature[idx], WindSpeed: payload.Hourly.Wind[idx], Precipitation: payload.Hourly.Precip[idx], CloudCover: payload.Hourly.Cloud[idx], Raw: raw}, nil
}

func mapSnapshot(s models.WeatherSnapshot) WeatherResponse {
	return WeatherResponse{Temperature: s.Temperature, WindSpeed: s.WindSpeed, Precipitation: s.Precipitation, CloudCover: s.CloudCover, ConditionsLevel: s.ConditionsLevel, Score: s.Score, Explanation: explanationFor(s.WindSpeed, s.Precipitation, s.Temperature)}
}

func scoreWeather(temp, wind, precipitation float64, cloud int) (int, string, string) {
	score := 100
	if wind >= 8 {
		score -= 45
	} else if wind >= 6 {
		score -= 25
	} else if wind >= 4 {
		score -= 10
	}
	if precipitation > 0 {
		score -= 20
	} else {
		score += 5
	}
	if temp < 12 || temp > 32 {
		score -= 20
	} else if temp < 16 || temp > 28 {
		score -= 10
	} else {
		score += 5
	}
	if cloud > 90 {
		score -= 5
	}
	if score > 100 {
		score = 100
	}
	if score < 0 {
		score = 0
	}
	level := "Плохие"
	if score >= 80 {
		level = "Отличные"
	} else if score >= 60 {
		level = "Хорошие"
	} else if score >= 40 {
		level = "Нормальные"
	}
	return score, level, explanationFor(wind, precipitation, temp)
}

func explanationFor(wind, precipitation, temp float64) string {
	if wind >= 8 {
		return fmt.Sprintf("Ветер %.1f м/с → будет сложнее грести.", wind)
	}
	if precipitation > 0 {
		return "Есть осадки → возможен дискомфорт на маршруте."
	}
	if temp < 12 {
		return "Прохладно, рекомендуется гидрокостюм."
	}
	if temp > 32 {
		return "Жарко, обязательно вода и головной убор."
	}
	return "Условия комфортные для прогулки."
}
