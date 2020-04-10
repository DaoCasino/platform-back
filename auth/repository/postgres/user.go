package postgres

import (
	"context"
	"github.com/jackc/pgx/v4/pgxpool"
	"platform-backend/models"
)

const (
	selectUserCntByAccNameStmt = "SELECT count(*) FROM users WHERE account_name = $1"
	selectUserByAccNameStmt    = "SELECT * FROM users WHERE account_name = $1"
	insertUserStmt             = "INSERT INTO users VALUES ($1, $2)"
)

type User struct {
	AccountName string `db:"account_name"`
	Email       string `db:"email"`
}

type UserPostgresRepo struct {
	dbPool *pgxpool.Pool
}

func NewUserPostgresRepo(dbPool *pgxpool.Pool) *UserPostgresRepo {
	return &UserPostgresRepo{
		dbPool: dbPool,
	}
}

func (r *UserPostgresRepo) HasUser(ctx context.Context, accountName string) (bool, error) {
	conn, err := r.dbPool.Acquire(ctx)
	if err != nil {
		return false, err
	}

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

	user := new(User)
	err = conn.QueryRow(ctx, selectUserByAccNameStmt, accountName).Scan(
		&user.AccountName,
		&user.Email,
	)
	if err != nil {
		return nil, err
	}

	return toModelUser(user), nil
}

func (r *UserPostgresRepo) AddUser(ctx context.Context, user *models.User) error {
	conn, err := r.dbPool.Acquire(ctx)
	if err != nil {
		return err
	}

	_, err = conn.Exec(ctx, insertUserStmt, user.AccountName, user.Email)
	return err
}

func toPostgresUser(u *models.User) *User {
	return &User{
		AccountName: u.AccountName,
		Email:       u.Email,
	}
}

func toModelUser(u *User) *models.User {
	return &models.User{
		AccountName: u.AccountName,
		Email:       u.Email,
	}
}
