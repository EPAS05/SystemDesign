package models

import (
	"time"
)

type Unit struct {
	ID         int       `db:"id"`
	Name       string    `db:"name"`
	Multiplier float64   `db:"multiplier"`
	CreatedAt  time.Time `db:"created_at"`
	UpdatedAt  time.Time `db:"updated_at"`
}

type CreateUnitRequest struct {
	Name       string
	Multiplier float64
}

type UpdateUnitRequest struct {
	ID         int
	Name       string
	Multiplier float64
}

type SetUnitRequest struct {
	NodeId int
	UnitID *int
}

type SetDefaultUnitRequest struct {
	ProductID int
	UnitID    *int
}
