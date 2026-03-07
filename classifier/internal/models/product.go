package models

import (
	"time"
)

type Product struct {
	ID             int       `db:"id"`
	UnitType       *string   `db:"unit_type"`
	WeightPerMeter *float64  `db:"weight_per_meter"`
	PieceLength    *float64  `db:"piece_length"`
	DefaultUnitID  *int      `db:"default_unit_id"`
	CreatedAt      time.Time `db:"created_at"`
	UpdatedAt      time.Time `db:"updated_at"`
}

type CreateProductRequest struct {
	UnitType       *string
	WeightPerMeter *float64
	PieceLength    *float64
	DefaultUnitID  *int
}
