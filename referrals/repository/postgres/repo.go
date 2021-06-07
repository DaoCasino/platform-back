package postgres

import (
	"context"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

var (
	selectRefByAccNameStmt = "SELECT referral_id FROM referrals WHERE account_name = $1"
	insertRefStmt          = "INSERT into referrals VALUES ($1, $2)"
	countTotalReferredStmt = "SELECT count(account_name) FROM affiliates WHERE affiliate_id = $1"
)

type ReferralPostgresRepo struct {
	dbPool *pgxpool.Pool
}

func NewReferralPostgresRepo(dbPool *pgxpool.Pool) *ReferralPostgresRepo {
	return &ReferralPostgresRepo{dbPool: dbPool}
}

func (r *ReferralPostgresRepo) GetReferralID(ctx context.Context, accountName string) (string, error) {
	conn, err := r.dbPool.Acquire(ctx)
	if err != nil {
		return "", err
	}
	defer conn.Release()

	var refID string

	err = conn.QueryRow(ctx, selectRefByAccNameStmt, accountName).Scan(&refID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", nil
		}
		return "", err
	}

	return refID, nil
}

func (r *ReferralPostgresRepo) AddReferralID(ctx context.Context, accountName string, referralID string) error {
	conn, err := r.dbPool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	_, err = conn.Exec(ctx, insertRefStmt, accountName, referralID)

	return err
}

func (r *ReferralPostgresRepo) GetTotalReferred(ctx context.Context, referralID string) (int, error) {
	conn, err := r.dbPool.Acquire(ctx)
	if err != nil {
		return 0, err
	}
	defer conn.Release()

	var totalReferred int

	err = conn.QueryRow(ctx, countTotalReferredStmt, referralID).Scan(&totalReferred)
	if err != nil {
		return 0, err
	}

	return totalReferred, nil
}
