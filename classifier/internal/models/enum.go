package models

import (
	"time"
)

type Enum struct {
	ID          int       `db:"id"`
	Name        string    `db:"name"`
	Description *string   `db:"description"`
	Type        string    `db:"type"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

type CreateEnumRequest struct {
	Name        string
	Type        string
	Description *string
	ParentID    *int
}

type UpdateEnumRequest struct {
	ID          int
	Name        string
	Description *string
	Type        string
}

type EnumValue struct {
	ID        int       `db:"id"`
	EnumID    int       `db:"enum_id"`
	Value     string    `db:"value"`
	SortOrder int       `db:"sort_order"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type CreateEnumValueRequest struct {
	EnumID    int
	Value     string
	SortOrder *int
}

type UpdateEnumValueRequest struct {
	ID        int
	Value     string
	SortOrder *int
}

type ReorderEnumValuesRequest struct {
	EnumID   int
	ValueIDs []int
}
