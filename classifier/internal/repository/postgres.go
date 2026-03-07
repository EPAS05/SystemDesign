package repository

import (
	"classifier/internal/models"
	"context"
	"database/sql"
)

type PostgresRepository struct {
	db *sql.DB
}

var _ Repository = (*PostgresRepository)(nil)

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) getNodeByID(ctx context.Context, id int) (*models.Node, error) {
	query := `
		SELECT id, name, parent_id, node_type, is_terminal, unit_id, sort_order, unit_type, weight_per_meter, piece_length, default_unit_id,enum_id, created_at, updated_at		
		FROM classifier_nodes
		WHERE id = $1;
	`

	var node models.Node
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&node.ID,
		&node.Name,
		&node.ParentID,
		&node.NodeType,
		&node.IsTerminal,
		&node.UnitID,
		&node.SortOrder,
		&node.CreatedAt,
		&node.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &node, nil
}
