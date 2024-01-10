package main

import (
	"github.com/caselongo/user-registration-go/internal/config"
	ur "github.com/caselongo/user-registration-go/user-registration"
	"github.com/justinas/nosurf"
	"net/http"
)

// NoSurf is the csrf protection middleware
func NoSurf(next http.Handler) http.Handler {
	csrfHandler := nosurf.New(next)

	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true,
		Path:     "/",
		Secure:   app.InProduction,
		SameSite: http.SameSiteLaxMode,
	})

	return csrfHandler
}

// SessionLoad loads and saves session data for current request
func SessionLoad(next http.Handler) http.Handler {
	return session.LoadAndSave(next)
}

func Auth(next http.Handler) http.Handler {
	return checkAuth(false, "/login", next)
}

func NoAuth(next http.Handler) http.Handler {
	return checkAuth(true, "/", next)
}

func checkAuth(ok bool, url string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, okAuth := session.Get(r.Context(), config.KeyUser).(ur.User)
		if ok == okAuth {
			http.Redirect(w, r, url, http.StatusSeeOther)
			return
		}

		next.ServeHTTP(w, r)
	})
}
