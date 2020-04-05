package postgres

import (
	"context"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/randallmlough/pgxscan"
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

	Casino := new(Casino)
	row := conn.QueryRow(ctx, selectCasinoByIdStmt, casinoId)
	err = pgxscan.NewScanner(row).Scan(Casino)
	if err != nil {
		return nil, err
	}

	return toModelCasino(Casino), nil
}

func (r *CasinoPostgresRepo) AllCasinos(ctx context.Context) ([]*models.Casino, error) {
	conn, err := r.dbPool.Acquire(ctx)
	if err != nil {
		return nil, err
	}

	var dst []Casino

	rows, err := conn.Query(ctx, selectAllCasinosStmt)
	if err != nil {
		return nil, err
	}

	err = pgxscan.NewScanner(rows).Scan(&dst)

	ret := make([]*models.Casino, len(dst))
	for i, v := range dst {
		ret[i] = toModelCasino(&v)
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
