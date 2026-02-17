package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

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

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) { writeJSON(w, 200, map[string]any{"status": "ok"}) })
	mux.HandleFunc("/api/instructors", h.listInstructors)
	mux.HandleFunc("/api/instructors/", h.getInstructor)
	mux.HandleFunc("/api/routes", h.listRoutes)
	mux.HandleFunc("/api/availability", h.listAvailability)
	mux.HandleFunc("/api/weather", h.getWeather)
	mux.HandleFunc("/api/bookings", h.createBooking)
	mux.HandleFunc("/api/bookings/", h.getBooking)
	mux.HandleFunc("/api/admin/instructors", h.upsertInstructor)
	mux.HandleFunc("/api/admin/routes", h.upsertRoute)
	mux.HandleFunc("/api/admin/availability/bulk", h.bulkSlots)
	mux.HandleFunc("/api/admin/bookings/", h.patchBookingStatus)
}

func (h *Handler) listInstructors(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, 405, nil)
		return
	}
	minPrice, _ := strconv.Atoi(r.URL.Query().Get("min_price"))
	maxPrice, _ := strconv.Atoi(r.URL.Query().Get("max_price"))
	minRating, _ := strconv.ParseFloat(r.URL.Query().Get("min_rating"), 64)
	items, err := h.repo.ListInstructors(minPrice, maxPrice, minRating, r.URL.Query().Get("tag"))
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	writeJSON(w, 200, items)
}
func (h *Handler) getInstructor(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, 405, nil)
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/api/instructors/")
	item, err := h.repo.GetInstructor(id)
	if err != nil {
		writeErrMsg(w, 404, "not found")
		return
	}
	writeJSON(w, 200, item)
}
func (h *Handler) listRoutes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, 405, nil)
		return
	}
	items, err := h.repo.ListRoutes()
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	writeJSON(w, 200, items)
}
func (h *Handler) listAvailability(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, 405, nil)
		return
	}
	date, err := time.Parse("2006-01-02", r.URL.Query().Get("date"))
	if err != nil {
		writeErrMsg(w, 400, "invalid date")
		return
	}
	slots, err := h.repo.ListAvailability(date, r.URL.Query().Get("route_id"), r.URL.Query().Get("instructor_id"))
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	writeJSON(w, 200, slots)
}
func (h *Handler) getWeather(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, 405, nil)
		return
	}
	lat, err := strconv.ParseFloat(r.URL.Query().Get("lat"), 64)
	if err != nil {
		writeErrMsg(w, 400, "invalid lat")
		return
	}
	lng, err := strconv.ParseFloat(r.URL.Query().Get("lng"), 64)
	if err != nil {
		writeErrMsg(w, 400, "invalid lng")
		return
	}
	dt, err := time.Parse(time.RFC3339, r.URL.Query().Get("datetime"))
	if err != nil {
		writeErrMsg(w, 400, "invalid datetime")
		return
	}
	resp, err := h.weather.Get(lat, lng, dt, r.URL.Query().Get("route_id"), r.URL.Query().Get("instructor_id"))
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	writeJSON(w, 200, resp)
}
func (h *Handler) createBooking(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, 405, nil)
		return
	}
	var req models.Booking
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, 400, err)
		return
	}
	if req.CustomerName == "" || req.Phone == "" || req.Participants < 1 || req.SlotID == "" {
		writeErrMsg(w, 400, "missing required fields")
		return
	}
	if req.Options == nil {
		req.Options = map[string]any{}
	}
	if err := h.repo.CreateBooking(&req); err != nil {
		writeErrMsg(w, 400, "cannot book selected slot")
		return
	}
	writeJSON(w, 201, req)
}
func (h *Handler) getBooking(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, 405, nil)
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/api/bookings/")
	b, err := h.repo.GetBooking(id)
	if err != nil {
		writeErrMsg(w, 404, "not found")
		return
	}
	writeJSON(w, 200, b)
}
func (h *Handler) upsertInstructor(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodPut {
		writeJSON(w, 405, nil)
		return
	}
	var m models.Instructor
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		writeErr(w, 400, err)
		return
	}
	if err := h.repo.UpsertInstructor(&m); err != nil {
		writeErr(w, 500, err)
		return
	}
	writeJSON(w, 200, m)
}
func (h *Handler) upsertRoute(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodPut {
		writeJSON(w, 405, nil)
		return
	}
	var m models.Route
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		writeErr(w, 400, err)
		return
	}
	if err := h.repo.UpsertRoute(&m); err != nil {
		writeErr(w, 500, err)
		return
	}
	writeJSON(w, 200, m)
}
func (h *Handler) bulkSlots(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, 405, nil)
		return
	}
	var s []models.TimeSlot
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		writeErr(w, 400, err)
		return
	}
	if err := h.repo.BulkCreateSlots(s); err != nil {
		writeErr(w, 500, err)
		return
	}
	writeJSON(w, 201, map[string]any{"created": len(s)})
}
func (h *Handler) patchBookingStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		writeJSON(w, 405, nil)
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/api/admin/bookings/")
	id = strings.TrimSuffix(id, "/status")
	var req struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Status == "" {
		writeErrMsg(w, 400, "status required")
		return
	}
	if err := h.repo.PatchBookingStatus(id, req.Status); err != nil {
		writeErr(w, 500, err)
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true})
}

func writeErr(w http.ResponseWriter, code int, err error) {
	writeJSON(w, code, map[string]any{"error": err.Error()})
}
func writeErrMsg(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]any{"error": msg})
}
func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if v != nil {
		_ = json.NewEncoder(w).Encode(v)
	}
}
