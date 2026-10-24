package app

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"os"
	"strings"

	_ "github.com/lib/pq"
)

type App struct {
	ListenAddr            string
	OIDCClientID          string
	OIDCDiscoveryEndpoint string
	TMDBApikey            string

	DB         *sql.DB
	SigningKey []byte
}

func New() (*App, error) {
	tmdbApikey := os.Getenv("TMDB_APIKEY")
	if filename := os.Getenv("TMDB_APIKEY_FILE"); filename != "" {
		value, err := os.ReadFile(filename)
		if err != nil {
			return nil, fmt.Errorf("read TMDB_APIKEY_FILE: %w", err)
		}
		tmdbApikey = strings.TrimSpace(string(value))
	}

	signingKey := make([]byte, 32)
	rand.Read(signingKey)

	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		return nil, err
	}

	app := App{
		ListenAddr:            os.Getenv("LISTEN_ADDR"),
		OIDCClientID:          os.Getenv("OIDC_CLIENT_ID"),
		OIDCDiscoveryEndpoint: os.Getenv("OIDC_DISCOVERY_ENDPOINT"),
		TMDBApikey:            tmdbApikey,

		DB:         db,
		SigningKey: signingKey,
	}

	if app.ListenAddr == "" {
		app.ListenAddr = ":8080"
	}

	return &app, nil
}
