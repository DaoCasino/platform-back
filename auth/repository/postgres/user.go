package postgres

import (
	"context"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"platform-backend/db"
	"platform-backend/models"
	"strconv"
)

const (
	selectUserCntByAccNameStmt = "SELECT count(*) FROM users WHERE account_name = $1"
	selectUserByAccNameStmt    = "SELECT * FROM users WHERE account_name = $1"
	selectAffIDByNameStmt      = "SELECT affiliate_id FROM affiliates WHERE account_name = $1"
	insertUserStmt             = "INSERT INTO users VALUES ($1, $2)"
	insertAffiliateStmt        = "INSERT INTO affiliates VALUES ($1, $2)"
	updateUserTokenNonce       = "UPDATE users SET token_nonce = token_nonce + 1 WHERE account_name = $1"
	invalidateOldestSessions   = "DELETE FROM active_token_nonces WHERE id = (SELECT id FROM active_token_nonces WHERE account_name = $1 ORDER BY id ASC LIMIT 1)"
	insertActiveSession        = "INSERT INTO active_token_nonces (account_name, token_nonce) VALUES ($1, $2)"
	selectSessionsCnt          = "SELECT count(*) FROM active_token_nonces WHERE account_name = $1"
	selectSessionCnt           = "SELECT count(*) FROM active_token_nonces WHERE account_name = $1 AND token_nonce = $2"
	deleteOldSessions          = "DELETE FROM active_token_nonces WHERE created + $1 * INTERVAL '1 second' < current_timestamp"
	invalidateSession          = "DELETE FROM active_token_nonces WHERE account_name = $1 AND token_nonce = $2"
	deleteEmail                = "UPDATE users SET email = '' WHERE account_name = $1"
)

type User struct {
	AccountName string `db:"account_name"`
	Email       string `db:"email"`
	TokenNonce  int64  `db:"token_nonce"`
}

type UserPostgresRepo struct {
	dbPool          *pgxpool.Pool
	maxSessions     int64
	sessionLifetime int64
}

func NewUserPostgresRepo(dbPool *pgxpool.Pool, maxSessions int64, sessionLifetime int64) *UserPostgresRepo {
	return &UserPostgresRepo{
		dbPool:          dbPool,
		maxSessions:     maxSessions,
		sessionLifetime: sessionLifetime,
	}
}

func (r *UserPostgresRepo) HasUser(ctx context.Context, accountName string) (bool, error) {
	conn, err := r.dbPool.Acquire(ctx)
	if err != nil {
		return false, err
	}
	defer conn.Release()

	var cnt uint
	err = conn.QueryRow(ctx, selectUserCntByAccNameStmt, accountName).Scan(&cnt)
	if err != nil {
		return false, err
	}

	return cnt > 0, nil
}

func (r *UserPostgresRepo) GetUser(ctx context.Context, accountName string) (*models.User, error) {
	conn, err := r.dbPool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	user := new(User)
	err = conn.QueryRow(ctx, selectUserByAccNameStmt, accountName).Scan(
		&user.AccountName,
		&user.Email,
		&user.TokenNonce,
	)
	if err != nil {
		return nil, err
	}

	var affiliateID string
	err = conn.QueryRow(ctx, selectAffIDByNameStmt, accountName).Scan(&affiliateID)
	if err != nil && err != pgx.ErrNoRows {
		return nil, err
	}

	return toModelUser(user, affiliateID), nil
}

func (r *UserPostgresRepo) AddUser(ctx context.Context, user *models.User) error {
	conn, err := r.dbPool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	tx, err := conn.Begin(ctx)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, insertUserStmt, user.AccountName, user.Email)
	if err != nil {
		_ = tx.Rollback(ctx)
		return err
	}

	if user.AffiliateID != "" {
		_, err = tx.Exec(ctx, insertAffiliateStmt, user.AccountName, user.AffiliateID)
		if err != nil {
			_ = tx.Rollback(ctx)
			return err
		}
	}

	err = tx.Commit(ctx)
	return err
}

func (r *UserPostgresRepo) IsSessionActive(ctx context.Context, accountName string, nonce int64) (bool, error) {
	conn, err := r.dbPool.Acquire(ctx)
	if err != nil {
		return false, err
	}
	defer conn.Release()

	var cnt uint
	err = conn.QueryRow(ctx, selectSessionCnt, accountName, strconv.FormatInt(nonce, 10)).Scan(&cnt)
	if err != nil {
		return false, err
	}

	return cnt > 0, nil
}

func (r *UserPostgresRepo) InvalidateSession(ctx context.Context, accountName string, nonce int64) error {
	conn, err := db.DbPool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	_, err = conn.Exec(ctx, invalidateSession, accountName, strconv.FormatInt(nonce, 10))
	return err
}

func (r *UserPostgresRepo) InvalidateOldSessions(ctx context.Context) error {
	conn, err := db.DbPool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	_, err = conn.Exec(ctx, deleteOldSessions, r.sessionLifetime)
	return err
}

func (r *UserPostgresRepo) AddNewSession(ctx context.Context, accountName string) (int64, error) {
	conn, err := r.dbPool.Acquire(ctx)
	if err != nil {
		return 0, err
	}
	defer conn.Release()

	var cnt uint
	err = conn.QueryRow(ctx, selectSessionsCnt, accountName).Scan(&cnt)
	if err != nil {
		return 0, err
	}

	if cnt >= uint(r.maxSessions) {
		_, err = conn.Exec(ctx, invalidateOldestSessions, accountName)
		if err != nil {
			return 0, err
		}
	}

	_, err = conn.Exec(ctx, updateUserTokenNonce, accountName)
	if err != nil {
		return 0, err
	}

	user := User{}
	err = conn.QueryRow(ctx, selectUserByAccNameStmt, accountName).Scan(
		&user.AccountName,
		&user.Email,
		&user.TokenNonce,
	)
	if err != nil {
		return 0, err
	}

	_, err = conn.Exec(ctx, insertActiveSession, accountName, strconv.FormatInt(user.TokenNonce, 10))
	if err != nil {
		return 0, err
	}

	return user.TokenNonce, nil
}

func (r *UserPostgresRepo) DeleteEmail(ctx context.Context, accountName string) error {
	conn, err := db.DbPool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	_, err = conn.Exec(ctx, deleteEmail, accountName)
	return err
}

func toModelUser(u *User, affiliateID string) *models.User {
	return &models.User{
		AccountName: u.AccountName,
		Email:       u.Email,
		AffiliateID: affiliateID,
	}
}
