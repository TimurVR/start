package auth

import (
	"os"

	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
)

const (
	key    = "randomString"
	MaxAge = 86400
	IsProd = false
)

func NewAuth() {
	store := sessions.NewCookieStore([]byte(key))
	store.MaxAge(MaxAge)
	store.Options.Path = "/"
	store.Options.HttpOnly = true
	store.Options.Secure = IsProd

	gothic.Store = store
	goth.UseProviders(
		google.New(getEnv("GOOGLE_KEY"), getEnv("GOOGLE_SECRET"), "http://localhost:8080"),
	)
}

func getEnv(key string) string {
	value, exists := os.LookupEnv(key)
	if !exists || value == "" {
		return " "
	}
	return value
}
