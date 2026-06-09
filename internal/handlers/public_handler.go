package handlers

import "net/http"

func Home(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, []string{
		"web/templates/layouts/base.html",
		"web/templates/public/home.html",
	}, getTemplateData(r, nil))
}

func BookingPage(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, []string{
		"web/templates/layouts/base.html",
		"web/templates/public/booking.html",
	}, getTemplateData(r, nil))
}

func InstructorsPage(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, []string{
		"web/templates/layouts/base.html",
		"web/templates/public/instructors.html",
	}, getTemplateData(r, nil))
}

func Favicon(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}
