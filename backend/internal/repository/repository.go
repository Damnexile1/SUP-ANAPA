package repository

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sort"
	"strings"
	"sync"
	"time"

	"sup-anapa/backend/internal/models"
)

type Repository struct {
	mu       sync.RWMutex
	inst     map[string]models.Instructor
	routes   map[string]models.Route
	slots    map[string]models.TimeSlot
	bookings map[string]models.Booking
	weather  []models.WeatherSnapshot
}

func New() *Repository {
	r := &Repository{
		inst:     map[string]models.Instructor{},
		routes:   map[string]models.Route{},
		slots:    map[string]models.TimeSlot{},
		bookings: map[string]models.Booking{},
		weather:  []models.WeatherSnapshot{},
	}
	r.seed()
	return r
}

func id() string { b := make([]byte, 16); _, _ = rand.Read(b); return hex.EncodeToString(b) }

func (r *Repository) seed() {
	now := time.Now().UTC()
	i1 := models.Instructor{ID: "11111111111111111111111111111111", Name: "Алексей Морев", PhotoURL: "https://images.unsplash.com/photo-1500648767791-00dcc994a43e", Bio: "Спокойные прогулки для новичков и семей.", Rating: 4.9, ReviewsCount: 132, ExperienceYears: 7, Tags: []string{"новички", "дети", "закат"}, Languages: []string{"RU", "EN"}, BasePrice: 3000, IsActive: true, CreatedAt: now, UpdatedAt: now}
	i2 := models.Instructor{ID: "22222222222222222222222222222222", Name: "Мария Волна", PhotoURL: "https://images.unsplash.com/photo-1494790108377-be9c29b29330", Bio: "Тренировки и SUP-фитнес на реке.", Rating: 4.8, ReviewsCount: 96, ExperienceYears: 5, Tags: []string{"спорт", "новички"}, Languages: []string{"RU"}, BasePrice: 3200, IsActive: true, CreatedAt: now, UpdatedAt: now}
	r.inst[i1.ID] = i1
	r.inst[i2.ID] = i2
	r1 := models.Route{ID: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Title: "Река у Анапы — спокойная вода", DurationMinutes: 90, Difficulty: "easy", BasePrice: 2500, Description: "Идеально для первого SUP", LocationLat: 45.092, LocationLng: 37.268, LocationTitle: "Старт: река у Анапы", CreatedAt: now, UpdatedAt: now}
	r.routes[r1.ID] = r1
	for d := 0; d < 7; d++ {
		s := models.TimeSlot{ID: id(), InstructorID: i1.ID, RouteID: r1.ID, StartAt: time.Date(now.Year(), now.Month(), now.Day()+d, 9, 0, 0, 0, time.UTC), EndAt: time.Date(now.Year(), now.Month(), now.Day()+d, 10, 30, 0, 0, time.UTC), Capacity: 6, Remaining: 6, Status: "open", CreatedAt: now, UpdatedAt: now}
		r.slots[s.ID] = s
	}
}

func (r *Repository) ListInstructors(minPrice, maxPrice int, minRating float64, tag string) ([]models.Instructor, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := []models.Instructor{}
	for _, i := range r.inst {
		if !i.IsActive {
			continue
		}
		if minPrice > 0 && i.BasePrice < minPrice {
			continue
		}
		if maxPrice > 0 && i.BasePrice > maxPrice {
			continue
		}
		if minRating > 0 && i.Rating < minRating {
			continue
		}
		if tag != "" {
			ok := false
			for _, t := range i.Tags {
				if strings.EqualFold(t, tag) {
					ok = true
					break
				}
			}
			if !ok {
				continue
			}
		}
		out = append(out, i)
	}
	sort.Slice(out, func(a, b int) bool { return out[a].Rating > out[b].Rating })
	return out, nil
}
func (r *Repository) GetInstructor(id string) (models.Instructor, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	i, ok := r.inst[id]
	if !ok {
		return models.Instructor{}, errors.New("not found")
	}
	return i, nil
}
func (r *Repository) ListRoutes() ([]models.Route, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := []models.Route{}
	for _, v := range r.routes {
		out = append(out, v)
	}
	return out, nil
}
func (r *Repository) ListAvailability(date time.Time, routeID, instructorID string) ([]models.TimeSlot, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := []models.TimeSlot{}
	start := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	end := start.Add(24 * time.Hour)
	for _, s := range r.slots {
		if s.StartAt.Before(start) || !s.StartAt.Before(end) || s.Status != "open" || s.Remaining <= 0 {
			continue
		}
		if routeID != "" && s.RouteID != routeID {
			continue
		}
		if instructorID != "" && s.InstructorID != instructorID {
			continue
		}
		out = append(out, s)
	}
	sort.Slice(out, func(a, b int) bool { return out[a].StartAt.Before(out[b].StartAt) })
	return out, nil
}
func (r *Repository) CreateBooking(b *models.Booking) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	s, ok := r.slots[b.SlotID]
	if !ok || s.Status != "open" || s.Remaining < b.Participants {
		return errors.New("slot unavailable")
	}
	s.Remaining -= b.Participants
	if s.Remaining == 0 {
		s.Status = "closed"
	}
	s.UpdatedAt = time.Now().UTC()
	r.slots[s.ID] = s
	now := time.Now().UTC()
	b.ID = id()
	b.Status = "pending"
	b.CreatedAt = now
	b.UpdatedAt = now
	r.bookings[b.ID] = *b
	return nil
}
func (r *Repository) GetBooking(id string) (models.Booking, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	b, ok := r.bookings[id]
	if !ok {
		return models.Booking{}, errors.New("not found")
	}
	return b, nil
}
func (r *Repository) UpsertInstructor(item *models.Instructor) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if item.ID == "" {
		item.ID = id()
	}
	item.UpdatedAt = time.Now().UTC()
	if item.CreatedAt.IsZero() {
		item.CreatedAt = item.UpdatedAt
	}
	r.inst[item.ID] = *item
	return nil
}
func (r *Repository) UpsertRoute(item *models.Route) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if item.ID == "" {
		item.ID = id()
	}
	item.UpdatedAt = time.Now().UTC()
	if item.CreatedAt.IsZero() {
		item.CreatedAt = item.UpdatedAt
	}
	r.routes[item.ID] = *item
	return nil
}
func (r *Repository) BulkCreateSlots(slots []models.TimeSlot) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now().UTC()
	for _, s := range slots {
		if s.ID == "" {
			s.ID = id()
		}
		if s.Status == "" {
			s.Status = "open"
		}
		if s.Remaining == 0 {
			s.Remaining = s.Capacity
		}
		s.CreatedAt = now
		s.UpdatedAt = now
		r.slots[s.ID] = s
	}
	return nil
}
func (r *Repository) PatchBookingStatus(id, status string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	b, ok := r.bookings[id]
	if !ok {
		return errors.New("not found")
	}
	b.Status = status
	b.UpdatedAt = time.Now().UTC()
	r.bookings[id] = b
	return nil
}
func (r *Repository) FindWeatherSnapshot(lat, lng float64, timeFrom time.Time, ttl time.Duration) (models.WeatherSnapshot, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	threshold := time.Now().UTC().Add(-ttl)
	for i := len(r.weather) - 1; i >= 0; i-- {
		w := r.weather[i]
		if w.LocationLat == lat && w.LocationLng == lng && w.TimeFrom.Equal(timeFrom) && w.FetchedAt.After(threshold) {
			return w, nil
		}
	}
	return models.WeatherSnapshot{}, errors.New("not found")
}
func (r *Repository) SaveWeatherSnapshot(s *models.WeatherSnapshot) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if s.ID == "" {
		s.ID = id()
	}
	r.weather = append(r.weather, *s)
	return nil
}
func (r *Repository) SuggestedSlots(target time.Time, routeID, instructorID string, limit int) ([]models.TimeSlot, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := []models.TimeSlot{}
	for _, s := range r.slots {
		if s.Status != "open" || s.Remaining <= 0 {
			continue
		}
		if routeID != "" && s.RouteID != routeID {
			continue
		}
		if instructorID != "" && s.InstructorID != instructorID {
			continue
		}
		if s.StartAt.Before(target.Add(-4*time.Hour)) || s.StartAt.After(target.Add(8*time.Hour)) {
			continue
		}
		out = append(out, s)
	}
	sort.Slice(out, func(a, b int) bool {
		da := out[a].StartAt.Sub(target)
		if da < 0 {
			da = -da
		}
		db := out[b].StartAt.Sub(target)
		if db < 0 {
			db = -db
		}
		return da < db
	})
	if len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}
