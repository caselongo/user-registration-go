package main

import (
	"github.com/caselongo/user-registration-go/internal/config"
	"github.com/caselongo/user-registration-go/internal/handlers"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"net/http"
)

func routes(app *config.AppConfig) http.Handler {
	mux := chi.NewRouter()

	mux.Use(middleware.Recoverer)
	mux.Use(NoSurf)
	mux.Use(SessionLoad)

	mux.With(Auth).Get("/", handlers.Repo.Home)
	mux.With(NoAuth).Get("/login", handlers.Repo.Login)
	mux.Post("/login", handlers.Repo.PostLogin)
	mux.With(NoAuth).Get("/register", handlers.Repo.Register)
	mux.Post("/register", handlers.Repo.PostRegister)
	mux.With(NoAuth).Get("/confirm/{code}", handlers.Repo.Confirm)
	mux.With(NoAuth).Get("/forgot", handlers.Repo.Forgot)
	mux.Post("/forgot", handlers.Repo.PostForgot)
	mux.With(NoAuth).Get("/reset/{code}", handlers.Repo.Reset)
	mux.Post("/reset", handlers.Repo.PostReset)
	mux.With(Auth).Get("/logout", handlers.Repo.Logout)

	return mux
}
