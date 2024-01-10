package models

import (
	"github.com/caselongo/user-registration-go/internal/forms"
	ur "github.com/caselongo/user-registration-go/user-registration"
)

// TemplateData holds data sent from handlers to templates
type TemplateData struct {
	Data            map[string]interface{}
	CsrfToken       string
	Form            *forms.Form
	User            *ur.User
	IsAuthenticated bool
}
