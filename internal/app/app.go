package app

import (
	"database/sql"
	"fmt"

	"github.com/lagz0ne/remmd/internal/core"
	"github.com/lagz0ne/remmd/internal/store"
)

type App struct {
	DB            *sql.DB
	Docs          *store.DocumentRepo
	Links         *store.LinkRepo
	Subscriptions *store.SubscriptionRepo
	Relations     *store.RelationRepo
	Templates     *store.TemplateRepo
	Playbooks     *store.PlaybookStore
	Positions     *store.PositionStore
	Reviews       *core.ReviewService
	Snapshots     *store.SnapshotService
}

func New(dbPath string) (*App, error) {
	db, err := store.OpenDB(dbPath)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	if err := store.Migrate(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("run migrations: %w", err)
	}
	docs := store.NewDocumentRepo(db)
	links := store.NewLinkRepo(db)
	snapshots := store.NewSnapshotService(links, docs)
	reviews := core.NewReviewService(links, links, snapshots)

	return &App{
		DB:            db,
		Docs:          docs,
		Links:         links,
		Subscriptions: store.NewSubscriptionRepo(db),
		Relations:     store.NewRelationRepo(db),
		Templates:     store.NewTemplateRepo(db),
		Playbooks:     store.NewPlaybookStore(db),
		Positions:     store.NewPositionStore(db),
		Reviews:       reviews,
		Snapshots:     snapshots,
	}, nil
}

func (a *App) Close() error {
	if a.DB != nil {
		return store.CloseDB(a.DB)
	}
	return nil
}
