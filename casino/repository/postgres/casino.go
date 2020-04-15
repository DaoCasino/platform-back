package postgres

import (
	"context"
	"github.com/jackc/pgx/v4/pgxpool"
	"platform-backend/models"
)

const (
	selectCasinoByIdStmt = "SELECT * FROM Casinos WHERE id = $1"
	selectAllCasinosStmt = "SELECT * FROM Casinos"
)

type Casino struct {
	Id uint64 `db:"id"`
}

type CasinoPostgresRepo struct {
	dbPool *pgxpool.Pool
}

func NewCasinoPostgresRepo(dbPool *pgxpool.Pool) *CasinoPostgresRepo {
	return &CasinoPostgresRepo{
		dbPool: dbPool,
	}
}

func (r *CasinoPostgresRepo) GetCasino(ctx context.Context, casinoId uint64) (*models.Casino, error) {
	conn, err := r.dbPool.Acquire(ctx)
	if err != nil {
		return nil, err
	}

	casino := new(Casino)
	err = conn.QueryRow(ctx, selectCasinoByIdStmt, casinoId).Scan(
		&casino.Id,
	)
	if err != nil {
		return nil, err
	}

	return toModelCasino(casino), nil
}

func (r *CasinoPostgresRepo) AllCasinos(ctx context.Context) ([]*models.Casino, error) {
	conn, err := r.dbPool.Acquire(ctx)
	if err != nil {
		return nil, err
	}

	var ret []*models.Casino

	rows, err := conn.Query(ctx, selectAllCasinosStmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		casino := new(Casino)
		err := rows.Scan(
			&casino.Id,
		)
		if err != nil {
			return nil, err
		}
		ret = append(ret, toModelCasino(casino))
	}

	return ret, nil
}

func toPostgresCasino(u *models.Casino) *Casino {
	return &Casino{
		Id: u.Id,
	}
}

func toModelCasino(u *Casino) *models.Casino {
	return &models.Casino{
		Id: u.Id,
	}
}
