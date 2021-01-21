package postgres

import (
	"context"
	"github.com/jackc/pgx/v4/pgxpool"
)

var (
	SelectPaidCashbackByAccName = "SELECT paid_cashback FROM cashback where account_name = $1"
)

type CashbackPostgresRepo struct {
	dbPool *pgxpool.Pool
}

func NewCashbackPostgresRepo(dbPool *pgxpool.Pool) *CashbackPostgresRepo {
	return &CashbackPostgresRepo{dbPool: dbPool}
}

func (r *CashbackPostgresRepo) GetPaidCashback(ctx context.Context, accountName string) (float64, error) {
	conn, err := r.dbPool.Acquire(ctx)
	if err != nil {
		return 0, err
	}
	defer conn.Release()

	var paidCashback float64
	if err := conn.QueryRow(ctx, SelectPaidCashbackByAccName, accountName).Scan(&paidCashback); err != nil {
		return 0, err
	}

	return paidCashback, nil
}
