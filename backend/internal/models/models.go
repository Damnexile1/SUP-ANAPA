package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type Instructor struct {
	ID              uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	Name            string         `json:"name"`
	PhotoURL        string         `json:"photo_url"`
	Bio             string         `json:"bio"`
	Rating          float64        `json:"rating"`
	ReviewsCount    int            `json:"reviews_count"`
	ExperienceYears int            `json:"experience_years"`
	Tags            datatypes.JSON `gorm:"type:jsonb" json:"tags"`
	Languages       datatypes.JSON `gorm:"type:jsonb" json:"languages"`
	BasePrice       int            `json:"base_price"`
	IsActive        bool           `json:"is_active"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
}

type Route struct {
	ID              uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	Title           string    `json:"title"`
	DurationMinutes int       `json:"duration_minutes"`
	Difficulty      string    `json:"difficulty"`
	BasePrice       int       `json:"base_price"`
	Description     string    `json:"description"`
	LocationLat     float64   `json:"location_lat"`
	LocationLng     float64   `json:"location_lng"`
	LocationTitle   string    `json:"location_title"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type TimeSlot struct {
	ID           uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	InstructorID uuid.UUID `gorm:"type:uuid" json:"instructor_id"`
	RouteID      uuid.UUID `gorm:"type:uuid" json:"route_id"`
	StartAt      time.Time `json:"start_at"`
	EndAt        time.Time `json:"end_at"`
	Capacity     int       `json:"capacity"`
	Remaining    int       `json:"remaining"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Booking struct {
	ID           uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	InstructorID uuid.UUID      `gorm:"type:uuid" json:"instructor_id"`
	RouteID      uuid.UUID      `gorm:"type:uuid" json:"route_id"`
	SlotID       uuid.UUID      `gorm:"type:uuid" json:"slot_id"`
	CustomerName string         `json:"customer_name"`
	Phone        string         `json:"phone"`
	Messenger    string         `json:"messenger"`
	Participants int            `json:"participants"`
	Options      datatypes.JSON `gorm:"type:jsonb" json:"options"`
	PriceTotal   int            `json:"price_total"`
	Status       string         `json:"status"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}

type WeatherSnapshot struct {
	ID              uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	LocationLat     float64        `json:"location_lat"`
	LocationLng     float64        `json:"location_lng"`
	TimeFrom        time.Time      `json:"time_from"`
	TimeTo          time.Time      `json:"time_to"`
	Temperature     float64        `json:"temperature"`
	WindSpeed       float64        `json:"wind_speed"`
	Precipitation   float64        `json:"precipitation"`
	CloudCover      int            `json:"cloud_cover"`
	ConditionsLevel string         `json:"conditions_level"`
	Score           int            `json:"score"`
	Raw             datatypes.JSON `gorm:"type:jsonb" json:"raw"`
	FetchedAt       time.Time      `json:"fetched_at"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
}
