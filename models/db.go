package models

import (
	"context"
	"database/sql"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/jackc/pgx/v4/stdlib"
	"github.com/rs/zerolog/log"
	"platform-backend/config"
)

var dbPool *pgxpool.Pool

func migrateDatabase(pgxCfg *pgx.ConnConfig) error {
	connStr := stdlib.RegisterConnConfig(pgxCfg)
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		log.Fatal().Msgf("Database open error, %s", err.Error())
		return err
	}

	defer func() {
		if err := db.Close(); err != nil {
			log.Fatal().Msgf("Database close error, %s", err.Error())
		}
	}()

	var instance database.Driver
	instance, err = postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Fatal().Msgf("Migration driver creation error, %s", err.Error())
		return err
	}

	var m *migrate.Migrate
	m, err = migrate.NewWithDatabaseInstance("file://./migrations", pgxCfg.Database, instance)
	if err != nil {
		log.Fatal().Msgf("Migration creation error, %s", err.Error())
		return err
	}

	err = m.Up()
	if err == migrate.ErrNoChange {
		log.Info().Msgf("No change migrations")
		return nil
	}

	if err != nil {
		log.Fatal().Msgf("Migration error, %s", err.Error())
		return err
	}

	log.Debug().Msgf("Migration successful")

	return nil
}

func InitDB(ctx context.Context, config *config.DbConfig) error {
	poolCfg, err := pgxpool.ParseConfig(config.Url)
	if err != nil {
		log.Fatal().Msgf("Database parseConfig error, %s", err.Error())
		return err
	}

	poolCfg.MaxConns = config.MaxPoolConns
	poolCfg.MinConns = config.MinPoolConns

	if err := migrateDatabase(poolCfg.ConnConfig); err != nil {
		return err
	}

	pool, err := pgxpool.ConnectConfig(ctx, poolCfg)
	if err != nil {
		return err
	}
	dbPool = pool

	log.Info().Msgf("Database initialized")

	return nil
}
