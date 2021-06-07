package postgres

import (
	"context"
	"github.com/jackc/pgx/v4/pgxpool"
	"platform-backend/models"
)

const (
	selectPaidCashbackByAccNameStmt = "SELECT paid_cashback FROM cashback where account_name = $1"
	addUserStmt                     = "INSERT INTO cashback(account_name) VALUES ($1)"
	setEthAddrStmt                  = "UPDATE cashback SET eth_address = $2 WHERE account_name = $1"
	selectEthAddrStmt               = "SELECT eth_address from cashback WHERE account_name = $1"
	setStateClaimStmt               = "UPDATE cashback SET state = 'claim' WHERE account_name = $1"
	setStateAccruedStmt             = "UPDATE cashback SET state = 'accrued', paid_cashback = paid_cashback + $2 WHERE account_name = $1"
	fetchAllStmt                    = "SELECT account_name, eth_address, paid_cashback, state FROM cashback WHERE state = 'claim'"
	fetchOneStml                    = "SELECT account_name, eth_address, paid_cashback, state FROM cashback WHERE account_name = $1"
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
	if err := conn.QueryRow(ctx, selectPaidCashbackByAccNameStmt, accountName).Scan(&paidCashback); err != nil {
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

	_, err = conn.Exec(ctx, addUserStmt, accountName)
	return err
}

func (r *CashbackPostgresRepo) DeleteEthAddress(ctx context.Context, accountName string) error {
	conn, err := r.dbPool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	_, err = conn.Exec(ctx, setEthAddrStmt, accountName, nil)
	return err
}

func (r *CashbackPostgresRepo) SetEthAddress(ctx context.Context, accountName string, ethAddress string) error {
	conn, err := r.dbPool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	_, err = conn.Exec(ctx, setEthAddrStmt, accountName, ethAddress)
	return err
}

func (r *CashbackPostgresRepo) GetEthAddress(ctx context.Context, accountName string) (*string, error) {
	conn, err := r.dbPool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	ethAddr := new(string)

	err = conn.QueryRow(ctx, selectEthAddrStmt, accountName).Scan(&ethAddr)
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

	_, err = conn.Exec(ctx, setStateClaimStmt, accountName)
	return err
}

func (r *CashbackPostgresRepo) SetStateAccrued(ctx context.Context, accountName string, cashback float64) error {
	conn, err := r.dbPool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	_, err = conn.Exec(ctx, setStateAccruedStmt, accountName, cashback)
	return err
}

func (r *CashbackPostgresRepo) FetchAll(ctx context.Context) ([]*models.CashbackRow, error) {
	conn, err := r.dbPool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	rows, _ := conn.Query(ctx, fetchAllStmt)

	result := make([]*models.CashbackRow, 0)
	for rows.Next() {
		data := new(models.CashbackRow)
		err := rows.Scan(&data.AccountName, &data.EthAddress, &data.PaidCashback, &data.State)
		if err != nil {
			return nil, err
		}

		result = append(result, data)
	}
	return result, rows.Err()
}

func (r *CashbackPostgresRepo) FetchOne(ctx context.Context, accountName string) (*models.CashbackRow, error) {
	conn, err := r.dbPool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	data := new(models.CashbackRow)
	err = conn.QueryRow(ctx, fetchOneStml, accountName).Scan(&data.AccountName, &data.EthAddress, &data.PaidCashback, &data.State)
	if err != nil {
		return nil, err
	}
	return data, nil
}
