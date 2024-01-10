package config

import (
	"fmt"
	scs "github.com/alexedwards/scs/v2"
	user_registration "github.com/caselongo/user-registration-go/user-registration"
	"html/template"
	"log"
	"os"
)

const (
	defaultPort string = "8080"
	KeyUser     string = "user"
)

// AppConfig holds the application config
type AppConfig struct {
	port             string
	host             string
	isTest           bool
	UseCache         bool
	TemplateCache    map[string]*template.Template
	InfoLog          *log.Logger
	InProduction     bool
	Session          *scs.SessionManager
	UserRegistration *user_registration.UserRegistration
}

func NewApp() AppConfig {
	port := getPort()
	return AppConfig{
		port:   port,
		host:   getHost(port),
		isTest: isTest(),
	}
}

func getPort() string {
	var port = os.Getenv("PORT")

	// use a default port if there is nothing in the environment
	if port == "" {
		port = defaultPort
		fmt.Println("INFO: No PORT environment variable detected, defaulting to " + port)
	}

	return ":" + port
}

func getHost(port string) string {
	var host = os.Getenv("HOST")

	// if there is nothing in the environment we use localhost
	if host == "" {
		host = "http://localhost" + port
		fmt.Println("INFO: No HOST environment variable detected, defaulting to " + host)
	}

	return host
}

func isTest() bool {
	var env = os.Getenv("ENV")

	if env == "" {
		return true
	}

	return env != "LIVE"
}

func (a *AppConfig) Port() string {
	return a.port
}

func (a *AppConfig) Host() string {
	return a.host
}

func (a *AppConfig) IsTest() bool {
	return a.isTest
}
