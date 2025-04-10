package models

import (
	"github.com/google/uuid"
	"time"
)

type Server struct {
	Port        string        `yaml:"port"`
	Host        string        `yaml:"host"`
	timeout     time.Duration `yaml:"timeout"`
	idleTimeout time.Duration `yaml:"idle_timeout"`
}

type Database struct {
	Port     int    `yaml:"port"`
	Host     string `yaml:"host"`
	Password string `yaml:"password"`
	User     string `yaml:"user"`
	DBName   string `yaml:"dbname"`
}

type Wallet struct {
	WalletID uuid.UUID
	Balance  uint64
}

type WalletRequest struct {
	WalletID      uuid.UUID `json:"walletId"`
	Amount        int64     `json:"amount"`
	OperationType string    `json:"operationType"`
}
