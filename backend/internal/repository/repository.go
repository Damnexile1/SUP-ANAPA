package repository

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"sup-anapa/backend/internal/models"
)

type Repository struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Repository { return &Repository{db: db} }

func (r *Repository) ListInstructors(minPrice, maxPrice int, minRating float64, tag string) ([]models.Instructor, error) {
	q := r.db.Model(&models.Instructor{}).Where("is_active = true")
	if minPrice > 0 {
		q = q.Where("base_price >= ?", minPrice)
	}
	if maxPrice > 0 {
		q = q.Where("base_price <= ?", maxPrice)
	}
	if minRating > 0 {
		q = q.Where("rating >= ?", minRating)
	}
	if tag != "" {
		q = q.Where("tags::text ILIKE ?", "%"+tag+"%")
	}
	var items []models.Instructor
	return items, q.Order("rating desc").Find(&items).Error
}

func (r *Repository) GetInstructor(id string) (models.Instructor, error) {
	var item models.Instructor
	return item, r.db.First(&item, "id = ?", id).Error
}

func (r *Repository) ListRoutes() ([]models.Route, error) {
	var items []models.Route
	return items, r.db.Order("base_price asc").Find(&items).Error
}

func (r *Repository) ListAvailability(date time.Time, routeID, instructorID string) ([]models.TimeSlot, error) {
	start := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	end := start.Add(24 * time.Hour)
	q := r.db.Model(&models.TimeSlot{}).Where("start_at >= ? AND start_at < ? AND status = 'open' AND remaining > 0", start, end)
	if routeID != "" {
		q = q.Where("route_id = ?", routeID)
	}
	if instructorID != "" {
		q = q.Where("instructor_id = ?", instructorID)
	}
	var slots []models.TimeSlot
	return slots, q.Order("start_at asc").Find(&slots).Error
}

func (r *Repository) CreateBooking(booking *models.Booking) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var slot models.TimeSlot
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&slot, "id = ?", booking.SlotID).Error; err != nil {
			return err
		}
		if slot.Remaining < booking.Participants || slot.Status != "open" {
			return gorm.ErrInvalidData
		}
		slot.Remaining -= booking.Participants
		if slot.Remaining == 0 {
			slot.Status = "closed"
		}
		if err := tx.Save(&slot).Error; err != nil {
			return err
		}
		return tx.Create(booking).Error
	})
}

func (r *Repository) GetBooking(id string) (models.Booking, error) {
	var booking models.Booking
	return booking, r.db.First(&booking, "id = ?", id).Error
}

func (r *Repository) UpsertInstructor(item *models.Instructor) error {
	if item.ID == uuid.Nil {
		item.ID = uuid.New()
	}
	return r.db.Save(item).Error
}

func (r *Repository) UpsertRoute(item *models.Route) error {
	if item.ID == uuid.Nil {
		item.ID = uuid.New()
	}
	return r.db.Save(item).Error
}

func (r *Repository) BulkCreateSlots(slots []models.TimeSlot) error {
	for i := range slots {
		if slots[i].ID == uuid.Nil {
			slots[i].ID = uuid.New()
		}
	}
	return r.db.Create(&slots).Error
}

func (r *Repository) PatchBookingStatus(id, status string) error {
	return r.db.Model(&models.Booking{}).Where("id = ?", id).Update("status", status).Error
}

func (r *Repository) FindWeatherSnapshot(lat, lng float64, timeFrom time.Time, ttl time.Duration) (models.WeatherSnapshot, error) {
	var w models.WeatherSnapshot
	err := r.db.Where("location_lat = ? AND location_lng = ? AND time_from = ? AND fetched_at >= ?", lat, lng, timeFrom, time.Now().Add(-ttl)).First(&w).Error
	return w, err
}

func (r *Repository) SaveWeatherSnapshot(s *models.WeatherSnapshot) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return r.db.Create(s).Error
}

func (r *Repository) SuggestedSlots(target time.Time, routeID, instructorID string, limit int) ([]models.TimeSlot, error) {
	q := r.db.Model(&models.TimeSlot{}).Where("start_at >= ? AND start_at <= ? AND status='open' AND remaining>0", target.Add(-4*time.Hour), target.Add(8*time.Hour))
	if routeID != "" {
		q = q.Where("route_id = ?", routeID)
	}
	if instructorID != "" {
		q = q.Where("instructor_id = ?", instructorID)
	}
	var slots []models.TimeSlot
	return slots, q.Order("ABS(EXTRACT(EPOCH FROM (start_at - ?)))", target).Limit(limit).Find(&slots).Error
}
