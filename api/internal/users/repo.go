package users

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type Repo interface {
	GetByUserName(ctx context.Context, username string) (*User, error)
}

type repo struct {
	db *sqlx.DB
}

func NewRepo(db *sqlx.DB) (Repo, error) {
	return &repo{
		db: db,
	}, nil
}

func (r *repo) GetByUserName(ctx context.Context, username string) (*User, error) {
	query := `SELECT
			username,
			password
		FROM users
		WHERE username = ?`

	query = r.db.Rebind(query)
	var user User
	err := r.db.GetContext(ctx, &user, query, username)
	return &user, err
}
