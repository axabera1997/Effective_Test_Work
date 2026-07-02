package models

import (
	"log/slog"
	"time"
)

var (
	Logger         *slog.Logger
	MigrationsPath = "file://migrations"
	// MigrationsPath = "file://../../migrations"
	EnvPath = "./.env"
	DSN     = ""
)

type Subscription struct {
	Service_name string `json:"service_name"` // “Yandex Plus”,
	Price        int64  `json:"price"`        // “price”: 400,
	User_id      string `json:"user_id"`      // “user_id”: “60601fee-2bf1-4721-ae6f-7636e79a0cba”,
	Start_date   string `json:"start_date"`   // “start_date”: “07-2025”
	End_date     string `json:"end_date"`     // “start_date”: “07-2025”
	Sdt          time.Time
	Edt          time.Time
}

type RetStruct struct {
	Name string
	Cunt int64
}
