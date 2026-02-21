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
	ID         int       `db:"id"`
	Name       string    `db:"name"`
	ParentID   *int      `db:"parent_id"`
	NodeType   NodeType  `db:"node_type"`
	IsTerminal *bool     `db:"is_terminal"`
	UnitID     *int      `db:"unit_id"`
	SortOrder  int       `db:"sort_order"`
	CreatedAt  time.Time `db:"created_at"`
	UpdatedAt  time.Time `db:"updated_at"`
}

type CreateNodeRequest struct {
	Name       string
	ParentID   *int
	NodeType   NodeType
	IsTerminal *bool
	UnitID     *int
	SortOrder  *int
}

type SetNameRequest struct {
	NodeId int
	Name   string
}

type SetParentRequest struct {
	NodeId      int
	NewParentID *int
}
