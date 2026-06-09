package handlers

import "net/http"

func AdminLogin(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, []string{
		"web/templates/layouts/base.html",
		"web/templates/admin/admin-login.html",
	}, getTemplateData(r, nil))
}

func AdminLoginPost(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	admin, err := authSvc.Authenticate(r.Context(), username, password)
	if err != nil {
		data := getTemplateData(r, map[string]interface{}{
			"Error": "Неверный логин или пароль",
		})
		renderTemplate(w, []string{
			"web/templates/layouts/base.html",
			"web/templates/admin/admin-login.html",
		}, data)
		return
	}

	session, _ := store.Get(r, "admin-session")
	session.Values["admin_id"] = admin.ID
	session.Values["username"] = admin.Username
	if err := session.Save(r, w); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func AdminDashboard(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, []string{
		"web/templates/layouts/base.html",
		"web/templates/admin/admin-dashboard.html",
	}, getTemplateData(r, nil))
}

func AdminInstructors(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, []string{
		"web/templates/layouts/base.html",
		"web/templates/admin/admin-instructors.html",
	}, getTemplateData(r, nil))
}

func AdminSlots(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, []string{
		"web/templates/layouts/base.html",
		"web/templates/admin/admin-slots.html",
	}, getTemplateData(r, nil))
}

func AdminBookings(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, []string{
		"web/templates/layouts/base.html",
		"web/templates/admin/admin-bookings.html",
	}, getTemplateData(r, nil))
}

func AdminWalkTypes(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, []string{
		"web/templates/layouts/base.html",
		"web/templates/admin/admin-walk-types.html",
	}, getTemplateData(r, nil))
}

func AdminLogout(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "admin-session")
	session.Options.MaxAge = -1
	session.Save(r, w)
	http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
}
