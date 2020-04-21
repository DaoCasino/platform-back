package postgres

import "github.com/jackc/pgx/v4/pgxpool"

type GameSessionsPostgresRepo struct {
	dbPool *pgxpool.Pool
}

func NewGameSessionsPostgresRepo(dbPool *pgxpool.Pool) *GameSessionsPostgresRepo {
	return &GameSessionsPostgresRepo{
		dbPool: dbPool,
	}
}
