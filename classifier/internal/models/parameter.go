package models

import (
	"time"
)

type ParameterDefinition struct {
	ID            int       `db:"id"`
	ClassNodeID   int       `db:"class_node_id"`
	Name          string    `db:"name"`
	Description   *string   `db:"description"`
	ParameterType string    `db:"parameter_type"`
	UnitID        *int      `db:"unit_id"`
	EnumID        *int      `db:"enum_id"`
	IsRequired    bool      `db:"is_required"`
	SortOrder     int       `db:"sort_order"`
	CreatedAt     time.Time `db:"created_at"`
	UpdatedAt     time.Time `db:"updated_at"`
}

type ParameterConstraint struct {
	ID         int       `db:"id"`
	ParamDefID int       `db:"param_def_id"`
	MinValue   *float64  `db:"min_value"`
	MaxValue   *float64  `db:"max_value"`
	CreatedAt  time.Time `db:"created_at"`
	UpdatedAt  time.Time `db:"updated_at"`
}

type ParameterValue struct {
	ID           int       `db:"id"`
	ProductID    int       `db:"product_id"`
	ParamDefID   int       `db:"param_def_id"`
	ValueNumeric *float64  `db:"value_numeric"`
	ValueEnumID  *int      `db:"value_enum_id"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

type CreateParameterDefinitionRequest struct {
	ClassNodeID   int
	Name          string
	Description   *string
	ParameterType string
	UnitID        *int
	EnumID        *int
	IsRequired    bool
	SortOrder     *int
	Constraints   *ParameterConstraint
}

type UpdateParameterDefinitionRequest struct {
	ID          int
	Name        string
	Description *string
	UnitID      *int
	EnumID      *int
	IsRequired  bool
	SortOrder   *int
	Constraints *ParameterConstraint
}

type CreateParameterValueRequest struct {
	ProductID    int
	ParamDefID   int
	ValueNumeric *float64
	ValueEnumID  *int
}

type UpdateParameterValueRequest struct {
	ID           int
	ValueNumeric *float64
	ValueEnumID  *int
}

type ParameterFilter struct {
	ParamDefID int
	Operator   string
	Value      interface{}
}
