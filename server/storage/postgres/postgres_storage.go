package postgres

import (
	"context"
	"embed"
	"net/url"
	"os"

	"github.com/catalystsquad/app-utils-go/logging"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var migrations embed.FS

type PostgresStorage struct {
	db *sqlx.DB
}

func (p PostgresStorage) Initialize(ctx context.Context) (func(), error) {
	// connect to forge cockroachdb serverless instances
	err := loadCockroachRootCert(ctx)
	if err != nil {
		return nil, err
	}

	params := url.Values{}
	params.Set("sslrootcert", fn)
	params.Set("sslmode", "verify-full")

	connectionString := url.URL{
		Scheme:   "postgresql",
		User:     url.UserPassword(os.Getenv("DB_USERNAME"), os.Getenv("DB_PASSWORD")),
		Host:     os.Getenv("DB_HOST"),
		Path:     os.Getenv("DB_NAME"),
		RawQuery: params.Encode() + "&options=--cluster%3Dlanky-bird-5343", // options and clusert values need to remain un-encoded to connect:
	}
	logging.Log.Info("Connecting to postgres")
	sqlDb, err := sqlx.Connect("postgres", connectionString.String())
	if err != nil {
		return nil, err
	}

	// set goose file system to use the embedded migrations
	goose.SetBaseFS(migrations)
	logging.Log.Info("Running migrations")
	err = goose.Up(sqlDb.DB, "migrations")
	if err != nil {
		return nil, err
	}
	logging.Log.Info("connected to postgres")

	// return a function to close the connection when the application is shutting down
	return func() { sqlDb.Close() }, nil
}

func (p PostgresStorage) Ready(ctx context.Context) bool {
	return p.db.Ping() == nil
}
