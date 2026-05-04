package repository

import (
	"classifier/internal/models"
	"context"
	"database/sql"
	"errors"
)

func (r *PostgresRepository) CreateUnit(ctx context.Context, req models.CreateUnitRequest) (*models.Unit, error) {
	var unit models.Unit
	query := `
		INSERT INTO units (name, multiplier)
		VALUES ($1, $2)
		RETURNING id, name, multiplier, created_at, updated_at
	`
	err := r.db.QueryRowContext(
		ctx,
		query,
		req.Name,
		req.Multiplier,
	).Scan(
		&unit.ID,
		&unit.Name,
		&unit.Multiplier,
		&unit.CreatedAt,
		&unit.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &unit, nil
}

func (r *PostgresRepository) GetUnit(ctx context.Context, id int) (*models.Unit, error) {
	var unit models.Unit
	query := `
		SELECT id, name, multiplier, created_at, updated_at 
		FROM units 
		WHERE id = $1;
		`
	err := r.db.QueryRowContext(ctx,
		query,
		id,
	).Scan(
		&unit.ID,
		&unit.Name,
		&unit.Multiplier,
		&unit.CreatedAt,
		&unit.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &unit, nil
}

func (r *PostgresRepository) GetAllUnits(ctx context.Context) ([]*models.Unit, error) {
	query := `
		SELECT id, name, multiplier, created_at, updated_at 
		FROM units 
		ORDER BY id;
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var units []*models.Unit
	for rows.Next() {
		var u models.Unit
		if err := rows.Scan(&u.ID, &u.Name, &u.Multiplier, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		units = append(units, &u)
	}
	return units, rows.Err()
}

func (r *PostgresRepository) UpdateUnit(ctx context.Context, req models.UpdateUnitRequest) (*models.Unit, error) {
	var unit models.Unit
	query := `
		UPDATE units 
		SET name = $1, multiplier = $2, updated_at = now() 
		WHERE id = $3
		RETURNING id, name, multiplier, created_at, updated_at;
	`
	err := r.db.QueryRowContext(ctx, query, req.Name, req.Multiplier, req.ID).Scan(
		&unit.ID,
		&unit.Name,
		&unit.Multiplier,
		&unit.CreatedAt,
		&unit.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &unit, nil
}

func (r *PostgresRepository) SetUnit(ctx context.Context, req models.SetUnitRequest) (*models.Node, error) {
	if req.UnitID != nil {
		if _, err := r.GetUnit(ctx, *req.UnitID); err != nil {
			return nil, err
		}
	}

	query := `
		UPDATE classifier_nodes
		SET unit_id = $1, updated_at = now()
		WHERE id = $2
		RETURNING id, name, parent_id, is_terminal, unit_id, sort_order, created_at, updated_at
	`

	var result models.Node
	err := r.db.QueryRowContext(ctx, query, req.UnitID, req.NodeId).Scan(
		&result.ID,
		&result.Name,
		&result.ParentID,
		&result.IsTerminal,
		&result.UnitID,
		&result.SortOrder,
		&result.CreatedAt,
		&result.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (r *PostgresRepository) SetDefaultUnit(ctx context.Context, req models.SetDefaultUnitRequest) (*models.Product, error) {
	if req.UnitID != nil {
		if _, err := r.GetUnit(ctx, *req.UnitID); err != nil {
			return nil, err
		}
	}

	var product models.Product
	query := `
		UPDATE products
		SET default_unit_id = $1, updated_at = now()
		WHERE id = $2
		RETURNING id, name, class_node_id, unit_type, weight_per_meter, piece_length, default_unit_id, created_at, updated_at
	`
	err := r.db.QueryRowContext(ctx, query, req.UnitID, req.ProductID).Scan(
		&product.ID,
		&product.Name,
		&product.ClassNodeID,
		&product.UnitType,
		&product.WeightPerMeter,
		&product.PieceLength,
		&product.DefaultUnitID,
		&product.CreatedAt,
		&product.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &product, nil
}

func (r *PostgresRepository) DeleteUnit(ctx context.Context, id int) error {
	var count int
	QueryToCheck := `
		SELECT COUNT(*) 
		FROM classifier_nodes 
		WHERE unit_id = $1;
	`
	err := r.db.QueryRowContext(ctx, QueryToCheck, id).Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New("unit is used by nodes")
	}
	QueryToDelete := `
		DELETE 
		FROM units 
		WHERE id = $1;
	`
	result, err := r.db.ExecContext(ctx, QueryToDelete, id)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}
