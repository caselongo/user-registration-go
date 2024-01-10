package render

import (
	"bytes"
	"github.com/caselongo/user-registration-go/internal/config"
	"github.com/caselongo/user-registration-go/internal/models"
	ur "github.com/caselongo/user-registration-go/user-registration"
	"github.com/justinas/nosurf"
	"html/template"
	"net/http"
	"path/filepath"
)

var functions = template.FuncMap{}

var app *config.AppConfig

// NewRenderer sets the config for the template package
func NewRenderer(a *config.AppConfig) {
	app = a
}

// AddDefaultData adds data for all templates
func AddDefaultData(td *models.TemplateData, r *http.Request) error {
	td.CsrfToken = nosurf.Token(r)

	if td.Data == nil {
		td.Data = make(map[string]interface{})
	}

	user, ok := app.Session.Get(r.Context(), config.KeyUser).(ur.User)
	if ok {
		td.User = &user
		td.IsAuthenticated = true
	} else {
		td.IsAuthenticated = false
	}

	return nil
}

// RenderTemplate renders a template
func RenderTemplate(w http.ResponseWriter, r *http.Request, tmpl string, td *models.TemplateData) {
	var tc map[string]*template.Template

	if app.UseCache {
		// get the template cache from the app config
		tc = app.TemplateCache
	} else {
		tc, _ = CreateTemplateCache()
	}

	t, ok := tc[tmpl]
	if !ok {
		w.Write([]byte("Could not get template from template cache"))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	buf := new(bytes.Buffer)

	err := AddDefaultData(td, r)
	if err != nil {
		w.Write([]byte(err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = t.Execute(buf, td)
	if err != nil {
		w.Write([]byte(err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = buf.WriteTo(w)
	if err != nil {
		w.Write([]byte("error writing template to browser"))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// CreateTemplateCache creates a template cache as a map
func CreateTemplateCache() (map[string]*template.Template, error) {

	myCache := map[string]*template.Template{}

	pages, err := filepath.Glob("./templates/*.page.tmpl")
	if err != nil {
		return myCache, err
	}

	for _, page := range pages {
		name := filepath.Base(page)
		ts, err := template.New(name).Funcs(functions).ParseFiles(page)
		if err != nil {
			return myCache, err
		}

		matches, err := filepath.Glob("./templates/*.layout.tmpl")
		if err != nil {
			return myCache, err
		}

		if len(matches) > 0 {
			ts, err = ts.ParseGlob("./templates/*.layout.tmpl")
			if err != nil {
				return myCache, err
			}
		}

		myCache[name] = ts
	}

	blocks, err := filepath.Glob("./templates/*.block.tmpl")
	if err != nil {
		return myCache, err
	}

	for _, block := range blocks {
		name := filepath.Base(block)
		ts, err := template.New(name).Funcs(functions).ParseFiles(block)
		if err != nil {
			return myCache, err
		}

		myCache[name] = ts
	}

	return myCache, nil
}
