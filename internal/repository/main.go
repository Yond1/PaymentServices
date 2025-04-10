package repository

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"math"
	"math/rand"
	"paymentSystem/internal/database"
	"paymentSystem/internal/models"
	"time"
)

type Repository struct {
	db *database.Storage
}

type RepositoryInterface interface {
	GetBalance(id string) (int, error)
	ChangeBalance(models.Wallet) error
}

const (
	DEPOSIT  = "DEPOSIT"
	WITHDRAW = "WITHDRAW"
)

func NewRepository(storage *database.Storage) (*Repository, error) {
	return &Repository{
		db: storage,
	}, nil
}

func (r *Repository) GetBalance(id string) (int, error) {
	var balance int
	err := r.db.Postgres.QueryRow("SELECT balance FROM wallets WHERE wallet_id = $1", id).Scan(&balance)
	if err != nil {
		return 0, err
	}
	return balance, nil
}

func (r *Repository) ChangeBalance(ctx context.Context, walletId uuid.UUID, amount uint64, operationType string) error {
	const maxRetries = 15
	var lastErr error

	rand.Seed(time.Now().UnixNano())

	for i := 0; i < maxRetries; i++ {
		if i > 0 {
			delay := time.Duration(math.Pow(2, float64(i))) * 50 * time.Millisecond
			jitter := time.Duration(rand.Int63n(int64(delay / 2)))
			time.Sleep(delay + jitter)
		}

		txOpts := &sql.TxOptions{
			Isolation: sql.LevelReadCommitted,
			ReadOnly:  false,
		}

		tx, err := r.db.Postgres.BeginTx(ctx, txOpts)
		if err != nil {
			lastErr = fmt.Errorf("failed to begin transaction: %w", err)
			continue
		}

		var currentBalance uint64
		err = tx.QueryRowContext(ctx,
			"SELECT balance FROM wallets WHERE wallet_id = $1 FOR UPDATE SKIP LOCKED",
			walletId,
		).Scan(&currentBalance)
		if err != nil {
			tx.Rollback()
			lastErr = fmt.Errorf("select failed: %w", err)
			continue
		}

		newBalance := currentBalance
		switch operationType {
		case DEPOSIT:
			newBalance += amount
		case WITHDRAW:
			if currentBalance < amount {
				tx.Rollback()
				return fmt.Errorf("insufficient funds")
			}
			newBalance -= amount
		}

		_, err = tx.ExecContext(ctx,
			"UPDATE wallets SET balance = $1 WHERE wallet_id = $2",
			newBalance,
			walletId,
		)
		if err != nil {
			tx.Rollback()
			lastErr = fmt.Errorf("update failed: %w", err)
			continue
		}

		if err := tx.Commit(); err != nil {
			if isSerializationError(err) {
				lastErr = fmt.Errorf("serialization conflict: %w", err)
				continue
			}
			lastErr = fmt.Errorf("commit failed: %w", err)
			continue
		}

		return nil
	}

	return fmt.Errorf("transaction failed after %d retries: %w", maxRetries, lastErr)
}

func isSerializationError(err error) bool {
	if pgErr, ok := err.(*pq.Error); ok {
		return pgErr.Code == "40001"
	}
	return false
}
