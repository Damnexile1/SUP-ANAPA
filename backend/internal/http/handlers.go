package http

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"

	"sup-anapa/backend/internal/models"
	"sup-anapa/backend/internal/repository"
	"sup-anapa/backend/internal/service"
)

type Handler struct {
	repo    *repository.Repository
	weather *service.WeatherService
}

func NewHandler(repo *repository.Repository, weather *service.WeatherService) *Handler {
	return &Handler{repo: repo, weather: weather}
}

func (h *Handler) Register(r *gin.Engine) {
	r.GET("/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })
	api := r.Group("/api")
	{
		api.GET("/instructors", h.listInstructors)
		api.GET("/instructors/:id", h.getInstructor)
		api.GET("/routes", h.listRoutes)
		api.GET("/availability", h.listAvailability)
		api.GET("/weather", h.getWeather)
		api.POST("/bookings", h.createBooking)
		api.GET("/bookings/:id", h.getBooking)

		admin := api.Group("/admin")
		admin.POST("/instructors", h.upsertInstructor)
		admin.PUT("/instructors/:id", h.upsertInstructor)
		admin.POST("/routes", h.upsertRoute)
		admin.PUT("/routes/:id", h.upsertRoute)
		admin.POST("/availability/bulk", h.bulkSlots)
		admin.PATCH("/bookings/:id/status", h.patchBookingStatus)
	}
}

func (h *Handler) listInstructors(c *gin.Context) {
	minPrice, _ := strconv.Atoi(c.Query("min_price"))
	maxPrice, _ := strconv.Atoi(c.Query("max_price"))
	minRating, _ := strconv.ParseFloat(c.Query("min_rating"), 64)
	items, err := h.repo.ListInstructors(minPrice, maxPrice, minRating, c.Query("tag"))
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, items)
}
func (h *Handler) getInstructor(c *gin.Context) {
	item, err := h.repo.GetInstructor(c.Param("id"))
	if err != nil {
		c.JSON(404, gin.H{"error": "not found"})
		return
	}
	c.JSON(200, item)
}
func (h *Handler) listRoutes(c *gin.Context) {
	items, err := h.repo.ListRoutes()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, items)
}

func (h *Handler) listAvailability(c *gin.Context) {
	date, err := time.Parse("2006-01-02", c.Query("date"))
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid date"})
		return
	}
	slots, err := h.repo.ListAvailability(date, c.Query("route_id"), c.Query("instructor_id"))
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, slots)
}

func (h *Handler) getWeather(c *gin.Context) {
	lat, err := strconv.ParseFloat(c.Query("lat"), 64)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid lat"})
		return
	}
	lng, err := strconv.ParseFloat(c.Query("lng"), 64)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid lng"})
		return
	}
	datetime, err := time.Parse(time.RFC3339, c.Query("datetime"))
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid datetime"})
		return
	}
	resp, err := h.weather.Get(lat, lng, datetime, c.Query("route_id"), c.Query("instructor_id"))
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, resp)
}

func (h *Handler) createBooking(c *gin.Context) {
	var req models.Booking
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if req.CustomerName == "" || req.Phone == "" || req.Participants < 1 {
		c.JSON(400, gin.H{"error": "missing required fields"})
		return
	}
	req.Status = "pending"
	if len(req.Options) == 0 {
		req.Options = datatypes.JSON([]byte(`{}`))
	}
	if err := h.repo.CreateBooking(&req); err != nil {
		c.JSON(400, gin.H{"error": "cannot book selected slot"})
		return
	}
	c.JSON(201, req)
}

func (h *Handler) getBooking(c *gin.Context) {
	b, err := h.repo.GetBooking(c.Param("id"))
	if err != nil {
		c.JSON(404, gin.H{"error": "not found"})
		return
	}
	c.JSON(200, b)
}
func (h *Handler) upsertInstructor(c *gin.Context) {
	var m models.Instructor
	if err := c.ShouldBindJSON(&m); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if c.Param("id") != "" {
		m.ID = parseUUID(c.Param("id"))
	}
	if err := h.repo.UpsertInstructor(&m); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, m)
}
func (h *Handler) upsertRoute(c *gin.Context) {
	var m models.Route
	if err := c.ShouldBindJSON(&m); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if c.Param("id") != "" {
		m.ID = parseUUID(c.Param("id"))
	}
	if err := h.repo.UpsertRoute(&m); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, m)
}
func (h *Handler) bulkSlots(c *gin.Context) {
	var slots []models.TimeSlot
	if err := c.ShouldBindJSON(&slots); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if err := h.repo.BulkCreateSlots(slots); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(201, gin.H{"created": len(slots)})
}
func (h *Handler) patchBookingStatus(c *gin.Context) {
	var req struct {
		Status string `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Status == "" {
		c.JSON(400, gin.H{"error": "status required"})
		return
	}
	if err := h.repo.PatchBookingStatus(c.Param("id"), req.Status); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"ok": true})
}
