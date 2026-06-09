package handlers

import (
	"html/template"
	"log"
	"net/http"
	"sup-anapa/internal/repository"
	"sup-anapa/internal/services"
	bookinguc "sup-anapa/internal/usecase/booking"

	"github.com/gorilla/sessions"
)

var (
	store          *sessions.CookieStore
	authSvc        *services.AuthService
	userAuthSvc    *services.UserAuthService
	userRepo       *repository.UserRepository
	createBooking  *bookinguc.CreateBookingUseCase
	bookingRepo    *repository.BookingRepository
	instructorRepo *repository.InstructorRepository
	slotRepo       *repository.SlotRepository
)

func Init(sessionSecret string, authService *services.AuthService) {
	store = sessions.NewCookieStore([]byte(sessionSecret))
	authSvc = authService
}

func SetUserAuthService(userAuth *services.UserAuthService) {
	userAuthSvc = userAuth
}

func SetRepositories(booking *repository.BookingRepository, instructor *repository.InstructorRepository, slot *repository.SlotRepository) {
	bookingRepo = booking
	instructorRepo = instructor
	slotRepo = slot
}

func SetUserRepository(repo *repository.UserRepository) {
	userRepo = repo
}

func SetBookingUseCase(uc *bookinguc.CreateBookingUseCase) {
	createBooking = uc
}

func GetStore() *sessions.CookieStore {
	return store
}

func renderTemplate(w http.ResponseWriter, layoutFiles []string, data interface{}) {
	tmpl, err := template.ParseFiles(layoutFiles...)
	if err != nil {
		log.Printf("Error parsing template: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = tmpl.ExecuteTemplate(w, "base", data)
	if err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func getTemplateData(r *http.Request, data map[string]interface{}) map[string]interface{} {
	if data == nil {
		data = make(map[string]interface{})
	}

	session, err := store.Get(r, "admin-session")
	if err == nil {
		if username, ok := session.Values["username"].(string); ok {
			data["Username"] = username
		}
	}

	userSession, err := store.Get(r, "user-session")
	if err == nil {
		if username, ok := userSession.Values["username"].(string); ok {
			data["UserUsername"] = username
		}
	}

	return data
}
