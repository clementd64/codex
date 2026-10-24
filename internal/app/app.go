package app

import (
	"crypto/rand"
	"database/sql"
	"os"

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
		TMDBApikey:            os.Getenv("TMDB_APIKEY"),

		DB:         db,
		SigningKey: signingKey,
	}

	if app.ListenAddr == "" {
		app.ListenAddr = ":8080"
	}

	return &app, nil
}
