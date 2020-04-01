package models

import (
	"context"
	"github.com/randallmlough/pgxscan"
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

func HasUser(ctx context.Context, accountName string) (bool, error) {
	conn, err := dbPool.Acquire(ctx)
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

func GetUser(ctx context.Context, accountName string) (*User, error) {
	conn, err := dbPool.Acquire(ctx)
	if err != nil {
		return nil, err
	}

	user := new(User)
	row := conn.QueryRow(ctx, selectUserByAccNameStmt, accountName)
	err = pgxscan.NewScanner(row).Scan(user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func AddUser(ctx context.Context, user *User) error {
	conn, err := dbPool.Acquire(ctx)
	if err != nil {
		return err
	}

	_, err = conn.Exec(ctx, insertUserStmt, user.AccountName, user.Email)
	return err
}
