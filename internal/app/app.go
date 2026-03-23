package app

import (
	"database/sql"
	"fmt"

	"github.com/lagz0ne/remmd/internal/store"
)

// App is the composition root that wires all dependencies together.
type App struct {
	DB            *sql.DB
	Docs          *store.DocumentRepo
	Links         *store.LinkRepo
	Subscriptions *store.SubscriptionRepo
	Events        *store.EventStore
}

// New creates a new App, opening the database at dbPath and running migrations.
func New(dbPath string) (*App, error) {
	db, err := store.OpenDB(dbPath)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	if err := store.Migrate(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("run migrations: %w", err)
	}
	return &App{
		DB:            db,
		Docs:          store.NewDocumentRepo(db),
		Links:         store.NewLinkRepo(db),
		Subscriptions: store.NewSubscriptionRepo(db),
		Events:        store.NewEventStore(db),
	}, nil
}

// Close releases the database connection.
func (a *App) Close() error {
	if a.DB != nil {
		return store.CloseDB(a.DB)
	}
	return nil
}
