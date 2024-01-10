package handlers

import (
	"encoding/gob"
	"fmt"
	"github.com/caselongo/user-registration-go/internal/config"
	"github.com/caselongo/user-registration-go/internal/forms"
	"github.com/caselongo/user-registration-go/internal/models"
	"github.com/caselongo/user-registration-go/internal/render"
	ur "github.com/caselongo/user-registration-go/user-registration"
	"github.com/go-chi/chi"
	"net/http"
)

// Repo the repository used by the handlers
var Repo *Repository

// Repository is the repository type
type Repository struct {
	App *config.AppConfig
}

// NewHandlers creates a new handlers repository
func NewHandlers(a *config.AppConfig) {
	Repo = &Repository{
		App: a,
	}

	gob.Register(ur.User{})
}

func (m *Repository) Home(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, r, "home.page.tmpl", &models.TemplateData{})
}

func (m *Repository) Register(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, r, "register.page.tmpl", &models.TemplateData{
		Form: forms.New(nil),
	})
}

func (m *Repository) PostRegister(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		m.renderMessage(w, r, err.Error(), MessageStateDanger, false)
		return
	}

	form := forms.New(r.PostForm)
	data := make(map[string]interface{})

	renderPage := func(form *forms.Form) {
		render.RenderTemplate(w, r, "register.page.tmpl", &models.TemplateData{
			Form: form,
			Data: data,
		})
	}

	form.Required("email", "password", "confirm-password")
	form.IsEmail("email")

	data["email"] = r.FormValue("email")

	if !form.Valid() {
		renderPage(form)
		return
	}

	ok, errEmail, errPassword, errConfirmPassword, err := m.App.UserRegistration.Register(r.FormValue("email"), r.FormValue("password"), r.FormValue("confirm-password"))
	if err != nil {
		m.renderMessage(w, r, err.Error(), MessageStateDanger, false)
		return
	}

	if ok {
		if m.App.UserRegistration.HasMailSender() {
			m.renderMessage(w, r, fmt.Sprintf("A confirmation e-mail will be sent to %s. Please check your inbox.", r.FormValue("email")), MessageStateSuccess, false)
		} else {
			m.renderMessage(w, r, "Your have successfully been registered.", MessageStateSuccess, true)
		}
		return
	}

	if errEmail != "" {
		form.Errors.Add("email", errEmail)
	}

	if errPassword != "" {
		form.Errors.Add("password", errPassword)
	}

	if errConfirmPassword != "" {
		form.Errors.Add("confirm-password", errConfirmPassword)
	}

	renderPage(form)
	return
}

func (m *Repository) Login(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, r, "login.page.tmpl", &models.TemplateData{
		Form: forms.New(nil),
	})
}

func (m *Repository) PostLogin(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		m.renderMessage(w, r, err.Error(), MessageStateDanger, true)
		return
	}

	form := forms.New(r.PostForm)

	form.Required("email", "password")
	form.IsEmail("email")

	if !form.Valid() {
		data := make(map[string]interface{})
		data["email"] = r.FormValue("email")

		render.RenderTemplate(w, r, "login.page.tmpl", &models.TemplateData{
			Form: form,
			Data: data,
		})
		return
	}

	user, errEmail, errPassword, err := m.App.UserRegistration.Login(r.FormValue("email"), r.FormValue("password"))
	if err != nil {
		m.renderMessage(w, r, err.Error(), MessageStateDanger, true)
		return
	}

	if user != nil {
		m.App.Session.Put(r.Context(), config.KeyUser, user)

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}

	data := make(map[string]interface{})
	data["email"] = r.FormValue("email")

	if errEmail != "" {
		form.Errors.Add("email", errEmail)
	}

	if errPassword != "" {
		form.Errors.Add("password", errPassword)
	}

	render.RenderTemplate(w, r, "login.page.tmpl", &models.TemplateData{
		Form: form,
		Data: data,
	})
}

func (m *Repository) Logout(w http.ResponseWriter, r *http.Request) {
	err := m.App.Session.Destroy(r.Context())
	if err != nil {
		m.renderMessage(w, r, err.Error(), MessageStateDanger, true)
		return
	}
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (m *Repository) Confirm(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")

	err := m.App.UserRegistration.Confirm(code)
	if err != nil {
		m.renderMessage(w, r, err.Error(), MessageStateDanger, false)
		return
	}

	m.renderMessage(w, r, "Your have successfully been registered.", MessageStateSuccess, true)
}

func (m *Repository) Reset(w http.ResponseWriter, r *http.Request) {
	resetCode := chi.URLParam(r, "code")

	_, err := m.App.UserRegistration.ValidateResetCode(resetCode)
	if err != nil {
		m.renderMessage(w, r, err.Error(), MessageStateDanger, false)
		return
	}

	data := make(map[string]interface{})
	data["code"] = resetCode

	render.RenderTemplate(w, r, "reset.page.tmpl", &models.TemplateData{
		Form: forms.New(nil),
		Data: data,
	})
}

func (m *Repository) PostReset(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		m.renderMessage(w, r, err.Error(), MessageStateDanger, false)
		return
	}

	form := forms.New(r.PostForm)

	form.Required("password", "confirm-password")

	data := make(map[string]interface{})
	data["code"] = r.FormValue("code")

	if !form.Valid() {
		render.RenderTemplate(w, r, "reset.page.tmpl", &models.TemplateData{
			Form: form,
			Data: data,
		})
		return
	}

	ok, errPassword, errConfirmPassword, err := m.App.UserRegistration.Reset(r.FormValue("code"), r.FormValue("password"), r.FormValue("confirm-password"))
	if err != nil {
		m.renderMessage(w, r, err.Error(), MessageStateDanger, false)
		return
	}

	if ok {
		m.renderMessage(w, r, "Your new password has been saved.", MessageStateSuccess, true)
		return
	}

	if errPassword != "" {
		form.Errors.Add("password", errPassword)
	}

	if errConfirmPassword != "" {
		form.Errors.Add("confirm-password", errConfirmPassword)
	}

	render.RenderTemplate(w, r, "reset.page.tmpl", &models.TemplateData{
		Form: form,
		Data: data,
	})
}

func (m *Repository) Forgot(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, r, "forgot.page.tmpl", &models.TemplateData{
		Form: forms.New(nil),
	})
}

func (m *Repository) PostForgot(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		m.renderMessage(w, r, err.Error(), MessageStateDanger, false)
		return
	}

	form := forms.New(r.PostForm)

	form.Required("email")
	form.IsEmail("email")

	data := make(map[string]interface{})
	data["email"] = r.FormValue("email")

	if !form.Valid() {
		render.RenderTemplate(w, r, "forgot.page.tmpl", &models.TemplateData{
			Form: form,
			Data: data,
		})
		return
	}

	err = m.App.UserRegistration.Forgot(r.FormValue("email"))
	if err != nil {
		m.renderMessage(w, r, err.Error(), MessageStateDanger, false)
		return
	}

	m.renderMessage(w, r, fmt.Sprintf("A password reset e-mail will be sent to %s. Please check your inbox.", r.FormValue("email")), MessageStateSuccess, false)
}

type MessageState string

const (
	MessageStateSuccess MessageState = "success"
	MessageStateWarning MessageState = "warning"
	MessageStateDanger  MessageState = "danger"
)

func (m *Repository) renderMessage(w http.ResponseWriter, r *http.Request, message string, state MessageState, showLogin bool) {
	data := make(map[string]interface{})
	data["message"] = message
	data["state"] = string(state)
	data["show-login"] = showLogin

	render.RenderTemplate(w, r, "message.page.tmpl", &models.TemplateData{
		Data: data,
	})
}
