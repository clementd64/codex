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
	IGDBClientId          string
	IGDBClientSecret      string
	ListenAddr            string
	OIDCClientID          string
	OIDCDiscoveryEndpoint string
	TMDBApikey            string

	DB         *sql.DB
	SigningKey []byte
}

func New() (*App, error) {
	igdbClientId, err := envFile("IGDB_CLIENT_ID")
	if err != nil {
		return nil, err
	}
	igdbClientSecret, err := envFile("IGDB_CLIENT_SECRET")
	if err != nil {
		return nil, err
	}
	tmdbApikey, err := envFile("TMDB_APIKEY")
	if err != nil {
		return nil, err
	}

	signingKey := make([]byte, 32)
	rand.Read(signingKey)

	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		return nil, err
	}

	app := App{
		IGDBClientId:          igdbClientId,
		IGDBClientSecret:      igdbClientSecret,
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

func envFile(key string) (string, error) {
	value := os.Getenv(key)
	if filename := os.Getenv(key + "_FILE"); filename != "" {
		file, err := os.ReadFile(filename)
		if err != nil {
			return "", fmt.Errorf("read %s_FILE: %w", key, err)
		}
		value = strings.TrimSpace(string(file))
	}
	return value, nil
}
