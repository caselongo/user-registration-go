package main

import (
	"fmt"
	"github.com/alexedwards/scs/v2"
	"github.com/caselongo/user-registration-go/internal/config"
	"github.com/caselongo/user-registration-go/internal/handlers"
	"github.com/caselongo/user-registration-go/internal/render"
	ur "github.com/caselongo/user-registration-go/user-registration"
	"log"
	"net/http"
	"time"
)

var app = config.NewApp()
var session *scs.SessionManager

// main is the main function
func main() {
	app.InProduction = !app.IsTest()

	// set up the session
	session = scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = app.InProduction

	app.Session = session

	tc, err := render.CreateTemplateCache()
	if err != nil {
		log.Fatalf("cannot create template cache: %s\n ", err.Error())
	}
	app.TemplateCache = tc

	app.UseCache = false

	handlers.NewHandlers(&app)
	render.NewRenderer(&app)

	log.Println("IsTest =", app.IsTest())

	uint1 := uint(1)
	userSource := NewUserSource()
	mailSender := NewMailSender()
	userRegistration, err := ur.NewUserRegistration(&ur.NewUserRegistrationConfig{
		UserSource: userSource,
		MailSender: mailSender,
		PasswordRequirements: &ur.PasswordRequirements{
			MinUppers:   &uint1,
			MinNumbers:  &uint1,
			MinSpecials: &uint1,
		},
	})

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("starting mail listener...")
	mailSender.ListenForMail()
	defer mailSender.Close()

	app.UserRegistration = userRegistration

	srv := &http.Server{
		Addr:    app.Port(),
		Handler: routes(&app),
	}

	fmt.Println(fmt.Sprintf("Starting application on port %s", srv.Addr))

	err = srv.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
