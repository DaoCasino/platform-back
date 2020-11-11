package postgres

import (
	"context"
	"github.com/jackc/pgx/v4/pgxpool"
)

var (
	selectRefCntByAccNameStmt = "SELECT count(*) FROM referrals WHERE account_name = $1"
	selectRefByAccNameStmt    = "SELECT referral_id FROM referrals WHERE account_name = $1"
	insertRefStmt             = "INSERT into referrals VALUES ($1, $2)"
)

type ReferralPostgresRepo struct {
	dbPool *pgxpool.Pool
}

func NewReferralPostgresRepo(dbPool *pgxpool.Pool) *ReferralPostgresRepo {
	return &ReferralPostgresRepo{dbPool: dbPool}
}

func (r *ReferralPostgresRepo) HasReferralID(ctx context.Context, accountName string) (bool, error) {
	conn, err := r.dbPool.Acquire(ctx)
	if err != nil {
		return false, err
	}
	defer conn.Release()

	var cnt uint
	err = conn.QueryRow(ctx, selectRefCntByAccNameStmt, accountName).Scan(&cnt)
	if err != nil {
		return false, err
	}

	return cnt > 0, nil
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
