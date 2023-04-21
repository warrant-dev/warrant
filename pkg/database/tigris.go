package database

import (
	"context"

	"github.com/tigrisdata/tigris-client-go/tigris"

	"github.com/rs/zerolog/log"

	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/golang-migrate/migrate/v4/source/github"
	"github.com/warrant-dev/warrant/pkg/config"
)

type Tigris struct {
	T      *tigris.Client
	Config *config.TigrisConfig
}

func NewTigris(config *config.TigrisConfig) *Tigris {
	return &Tigris{
		Config: config,
	}
}

func (ds *Tigris) Type() string {
	return TypeTigris
}

func (ds *Tigris) Connect(ctx context.Context) error {
	t, err := tigris.NewClient(ctx,
		&tigris.Config{
			URL:          ds.Config.URL,
			ClientID:     ds.Config.ClientID,
			ClientSecret: ds.Config.ClientSecret,
			Project:      ds.Config.Project,
		})
	if err != nil {
		return err
	}

	ds.T = t

	log.Debug().Msgf("Connected to Tigris database %s", ds.Config.Project)

	return nil
}

func (ds *Tigris) Migrate(_ context.Context, version uint) error {
	log.Debug().Msgf("Migrating Tigris database %s", ds.Config.Project)
	return nil
}

func (ds *Tigris) Ping(ctx context.Context) error {
	return nil
}

func (ds *Tigris) WithinTransaction(ctx context.Context, txFunc func(ctx context.Context) error) error {
	panic("unimplemented")
}
