package repository

import (
	"context"
	"database/sql"
	"github.com/testcontainers/testcontainers-go"
	"paymentSystem/internal/database"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupTestDB(t *testing.T) *database.Storage {
	ctx := context.Background()
	container, err := postgres.RunContainer(
		ctx,
		testcontainers.WithImage("postgres:15"),
		postgres.WithDatabase("test_db"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second)),
	)
	if err != nil {
		t.Fatal(err)
	}

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}

	storage := &database.Storage{}
	db, err := sql.Open("postgres", connStr)
	db.SetMaxOpenConns(100)
	db.SetMaxIdleConns(50)
	db.SetConnMaxLifetime(10 * time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	storage.Postgres = db

	_, err = storage.Postgres.Exec("CREATE TABLE wallets (wallet_id UUID PRIMARY KEY, balance BIGINT)")
	if err != nil {
		t.Fatal(err)
	}

	return storage
}

func TestRepository_GetBalance(t *testing.T) {
	storage := setupTestDB(t)
	defer storage.Postgres.Close()

	repo, _ := NewRepository(storage)
	walletID := uuid.New()

	_, err := storage.Postgres.Exec(
		"INSERT INTO wallets (wallet_id, balance) VALUES ($1, $2)",
		walletID,
		1000,
	)
	assert.NoError(t, err)

	balance, err := repo.GetBalance(walletID.String())
	assert.NoError(t, err)
	assert.Equal(t, 1000, balance)
}

func TestRepository_ChangeBalance(t *testing.T) {
	storage := setupTestDB(t)
	defer storage.Postgres.Close()

	repo, _ := NewRepository(storage)
	walletID := uuid.New()

	_, err := storage.Postgres.Exec(
		"INSERT INTO wallets (wallet_id, balance) VALUES ($1, $2)",
		walletID,
		1000,
	)
	assert.NoError(t, err)

	err = repo.ChangeBalance(context.Background(), walletID, 500, "DEPOSIT")
	assert.NoError(t, err)

	var balance int64
	err = storage.Postgres.QueryRow("SELECT balance FROM wallets WHERE wallet_id = $1", walletID).Scan(&balance)
	assert.NoError(t, err)
	assert.Equal(t, int64(1500), balance)

	err = repo.ChangeBalance(context.Background(), walletID, 2000, "WITHDRAW")
	assert.ErrorContains(t, err, "not enough money")
}

func TestConcurrentRequests(t *testing.T) {
	storage := setupTestDB(t)
	defer storage.Postgres.Close()

	repo, _ := NewRepository(storage)
	walletID := uuid.New()

	_, err := storage.Postgres.Exec(
		"INSERT INTO wallets (wallet_id, balance) VALUES ($1, $2)",
		walletID,
		0,
	)
	assert.NoError(t, err)

	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		time.Sleep(time.Millisecond * 1)
		go func() {
			defer wg.Done()
			err := repo.ChangeBalance(context.Background(), walletID, 1, "DEPOSIT")
			if err != nil {
				t.Logf("Transaction error: %v", err)
			}
		}()
	}
	wg.Wait()

	var balance int64
	err = storage.Postgres.QueryRow("SELECT balance FROM wallets WHERE wallet_id = $1", walletID).Scan(&balance)
	assert.NoError(t, err)
	assert.Equal(t, int64(1000), balance)
}
