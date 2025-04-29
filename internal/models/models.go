package models

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type Order struct {
	ID         int64
	UserID     int64
	Number     string
	Status     string
	Accrual    pgtype.Float8
	UploadedAt pgtype.Timestamptz
}

type User struct {
	ID        int64
	Login     string
	Password  string
	Balance   pgtype.Float8
	Withdrawn pgtype.Float8
}

type Withdrawal struct {
	UserID      int64
	OrderNumber string
	Sum         pgtype.Float8
	ProcessedAt pgtype.Timestamptz
}
