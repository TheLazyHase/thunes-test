package transfers

import (
	"context"
	"fmt"

	"thunes-api/internal/users"

	"github.com/jmoiron/sqlx"
)

type Repo interface {
	GetByUser(ctx context.Context, username users.User) ([]Transfer, error)
	CreateTransfer(ctx context.Context, transfer Transfer) (*Transfer, error)
}

type repo struct {
	db *sqlx.DB
}

func NewRepo(db *sqlx.DB) (Repo, error) {
	return &repo{
		db: db,
	}, nil
}

func (r *repo) GetByUser(ctx context.Context, user users.User) ([]Transfer, error) {
	query := `SELECT
			source_id,
			destination_id,
			amount,
			execution_date
		FROM transfers
		WHERE source_id = ? OR destination_id = ?`

	query = r.db.Rebind(query)
	var transfers []Transfer
	err := r.db.SelectContext(ctx, &transfers, query, user.ID, user.ID)

	return transfers, err
}

func (r *repo) GetBalanceDuringTransaction(ctx context.Context, tx *sqlx.Tx, ID int64) (int64, error) {
	query := `SELECT
			balance
		FROM users
		WHERE id = ?`

	query = r.db.Rebind(query)
	var balance int64
	err := r.db.GetContext(ctx, &balance, query, ID)
	if err != nil {
		return 0, fmt.Errorf("Error retrieving balance: %w", err)
	}
	return balance, nil
}

func (r *repo) CreateTransfer(ctx context.Context, requestedTransfer Transfer) (*Transfer, error) {
	//Transaction, so everything is done atomically
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("Error starting sqlx transaction: %w", err)
	}

	//check balance - this is done with transaction-locked balance to avoid issues
	balance, err := r.GetBalanceDuringTransaction(ctx, tx, requestedTransfer.SourceID)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("error checking balance: %w", err)
	}
	if balance < requestedTransfer.Amount {
		tx.Rollback()
		return nil, fmt.Errorf("Not enough funds: %d vs %d", balance, requestedTransfer.Amount)
	}

	//Create the actual record
	query := `INSERT INTO transfers
		(source_id, destination_id, amount, execution_date) VALUES (?, ?, ?, NOW())
		RETURNING source_id, destination_id, amount, execution_date`

	query = r.db.Rebind(query)
	var transfer Transfer
	err = tx.GetContext(ctx, &transfer, query, requestedTransfer.SourceID, requestedTransfer.DestinationID, requestedTransfer.Amount)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("Error inserting transfer: %w", err)
	}

	//change balance for source
	query = `UPDATE users
		SET balance = balance + :balance
		WHERE id = :id`

	_, err = tx.NamedExecContext(ctx,
		query,
		map[string]interface{}{
			"id": requestedTransfer.SourceID,
			"balance": -requestedTransfer.Amount,
		},
	)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("Error changing source balance: %w", err)
	}

	//change balance for destination
	query = `UPDATE users
		SET balance = balance + :balance
		WHERE id = :id`

	_, err = tx.NamedExecContext(ctx,
		query,
		map[string]interface{}{
			"id": requestedTransfer.DestinationID,
			"balance": requestedTransfer.Amount,
		},
	)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("Error changing destination balance: %w", err)
	}

	//Commit transaction
	err = tx.Commit()
	if err != nil {
		return nil, fmt.Errorf("Error commiting sqlx transaction: %w", err)
	}

	return &transfer, nil
}
