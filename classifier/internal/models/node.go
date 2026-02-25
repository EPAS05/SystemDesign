package models

import (
	"time"
)

type NodeType string

const (
	TypeMetaclass NodeType = "metaclass"
	TypeLeaf      NodeType = "leaf"
)

type Node struct {
	ID             int       `db:"id"`
	Name           string    `db:"name"`
	ParentID       *int      `db:"parent_id"`
	NodeType       NodeType  `db:"node_type"`
	IsTerminal     *bool     `db:"is_terminal"`
	UnitID         *int      `db:"unit_id"`
	SortOrder      int       `db:"sort_order"`
	UnitType       *string   `db:"unit_type"`
	WeightPerMeter *float64  `db:"weight_per_meter"`
	PieceLength    *float64  `db:"piece_length"`
	DefaultUnitID  *int      `db:"default_unit_id"`
	CreatedAt      time.Time `db:"created_at"`
	UpdatedAt      time.Time `db:"updated_at"`
}

type CreateNodeRequest struct {
	Name           string
	ParentID       *int
	NodeType       NodeType
	IsTerminal     *bool
	UnitID         *int
	SortOrder      *int
	UnitType       *string
	WeightPerMeter *float64
	PieceLength    *float64
	DefaultUnitID  *int
}

type SetNameRequest struct {
	NodeId int
	Name   string
}

type SetParentRequest struct {
	NodeId      int
	NewParentID *int
}
