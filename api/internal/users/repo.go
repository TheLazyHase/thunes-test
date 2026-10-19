package users

import (
	"context"
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
)

type Repo interface {
	GetByUserName(ctx context.Context, username string) (*User, error)
	GetByID(ctx context.Context, id int64) (*User, error)
	GetByIDs(ctx context.Context, ids []int64) (map[int64]User, error)
	GetAll(ctx context.Context) ([]User, error)
	//This is for testing, not for actual transactions !
	SetBalance(ctx context.Context, user *User, balance int64) error
}

type repo struct {
	db *sqlx.DB
}

func NewRepo(db *sqlx.DB) (Repo, error) {
	return &repo{
		db: db,
	}, nil
}

//TODO: hash the passwords
func (r *repo) SetBalance(ctx context.Context, user *User, balance int64) error {
	query := `UPDATE users
		SET balance = :balance
		WHERE id = :id`

	_, err := r.db.NamedExecContext(
		ctx,
		query,
		map[string]interface{}{
			"id": user.ID,
			"balance": balance,
		},
	)
	if err != nil {
		return fmt.Errorf("Error changing balance: %w", err)
	}
	return nil
}

func (r *repo) GetByUserName(ctx context.Context, username string) (*User, error) {
	query := `SELECT
			id,
			username,
			password,
			balance
		FROM users
		WHERE username = ?`

	query = r.db.Rebind(query)
	var user User
	err := r.db.GetContext(ctx, &user, query, username)
	if err != nil {
		return &user, fmt.Errorf("Error retrieving users by username: %w", err)
	}
	return &user, nil
}

func (r *repo) GetByID(ctx context.Context, id int64) (*User, error) {
	query := `SELECT
			id,
			username,
			password,
			balance
		FROM users
		WHERE id = ?`

	query = r.db.Rebind(query)
	var user User
	err := r.db.GetContext(ctx, &user, query, id)
	if err != nil {
		return &user, fmt.Errorf("Error retrieving user by id: %w", err)
	}
	return &user, nil
}

func (r *repo) GetAll(ctx context.Context) ([]User, error) {
	//We don't retrieve balance nor passwords (or hashes) when retrieving multiple users
	query := `SELECT
			id,
			username
		FROM users`

	query = r.db.Rebind(query)
	var users []User
	err := r.db.SelectContext(ctx, &users, query)
	if err != nil {
		return users, fmt.Errorf("Error retrieving users: %w", err)
	}
	return users, nil
}

func (r *repo) GetByIDs(ctx context.Context, ids []int64) (map[int64]User, error) {
	//We don't retrieve balance nor passwords (or hashes) when retrieving multiple users
	query, args, err := sqlx.In(`SELECT
			id,
			username
		FROM users
		WHERE id IN (?)`, ids)
	if err != nil {
		return nil, fmt.Errorf("Error hydrating query: %w", err)
	}
	query = r.db.Rebind(query)

	var users []User
	log.Printf("what is query ? %s %v", query, args)
	err = r.db.SelectContext(ctx, &users, query, args...)
	if err != nil {
		return nil, fmt.Errorf("Error retrieving users: %w", err)
	}
	userMap := make(map[int64]User)
	for _, u := range users {
		userMap[u.ID] = u
	}
	return userMap, nil
}
