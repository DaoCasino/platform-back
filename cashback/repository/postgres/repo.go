package postgres

import (
	"context"
	"github.com/jackc/pgx/v4/pgxpool"
)

var (
	SelectPaidCashbackByAccNameStmt = "SELECT paid_cashback FROM cashback where account_name = $1"
	AddUserStmt                     = "INSERT INTO cashback(account_name) VALUES ($1)"
	SetEthAddrStmt                  = "UPDATE cashback SET eth_address = $2 WHERE account_name = $1"
	SelectEthAddrStmt               = "SELECT eth_address from cashback WHERE account_name = $1"
	SetStateClaimStmt               = "UPDATE cashback SET state = 'claim' WHERE account_name = $1"
	SetStateAccruedStmt             = "UPDATE cashback SET state = 'accrued' WHERE account_name = $1"
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
	if err := conn.QueryRow(ctx, SelectPaidCashbackByAccNameStmt, accountName).Scan(&paidCashback); err != nil {
		return 0, err
	}

	return paidCashback, nil
}

func (r *CashbackPostgresRepo) AddUser(ctx context.Context, accountName string) error {
	conn, err := r.dbPool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	_, err = conn.Exec(ctx, AddUserStmt, accountName)
	return err
}

func (r *CashbackPostgresRepo) DeleteEthAddress(ctx context.Context, accountName string) error {
	conn, err := r.dbPool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	_, err = conn.Exec(ctx, SetEthAddrStmt, accountName, nil)
	return err
}

func (r *CashbackPostgresRepo) SetEthAddress(ctx context.Context, accountName string, ethAddress string) error {
	conn, err := r.dbPool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	_, err = conn.Exec(ctx, SetEthAddrStmt, accountName, ethAddress)
	return err
}

func (r *CashbackPostgresRepo) GetEthAddress(ctx context.Context, accountName string) (*string, error) {
	conn, err := r.dbPool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	ethAddr := new(string)

	err = conn.QueryRow(ctx, SelectEthAddrStmt, accountName).Scan(&ethAddr)
	if err != nil {
		return nil, err
	}

	return ethAddr, nil
}

func (r *CashbackPostgresRepo) SetStateClaim(ctx context.Context, accountName string) error {
	conn, err := r.dbPool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	_, err = conn.Exec(ctx, SetStateClaimStmt, accountName)
	return err
}

func (r *CashbackPostgresRepo) SetStateAccrued(ctx context.Context, accountName string) error {
	conn, err := r.dbPool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	_, err = conn.Exec(ctx, SetStateAccruedStmt, accountName)
	return err
}
