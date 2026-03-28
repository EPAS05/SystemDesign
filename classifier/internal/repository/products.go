package repository

import (
	"classifier/internal/models"
	"context"
	"database/sql"
)

func (r *PostgresRepository) CreateProduct(ctx context.Context, req models.CreateProductRequest) (*models.Product, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM classifier_nodes WHERE id = $1)`, req.ClassNodeID).Scan(&exists)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrNotFound
	}

	var product models.Product
	query := `
		INSERT INTO products (name, class_node_id, unit_type, weight_per_meter, piece_length, default_unit_id)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, name, class_node_id, unit_type, weight_per_meter, piece_length, default_unit_id, created_at, updated_at
	`
	err = r.db.QueryRowContext(ctx, query,
		req.Name,
		req.ClassNodeID,
		req.UnitType,
		req.WeightPerMeter,
		req.PieceLength,
		req.DefaultUnitID,
	).Scan(
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
	if err != nil {
		return nil, err
	}
	return &product, nil
}

func (r *PostgresRepository) GetProduct(ctx context.Context, id int) (*models.Product, error) {
	var product models.Product
	query := `
		SELECT id, name, class_node_id, unit_type, weight_per_meter, piece_length, default_unit_id, created_at, updated_at
		FROM products
		WHERE id = $1
	`
	err := r.db.QueryRowContext(ctx, query, id).Scan(
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

func (r *PostgresRepository) UpdateProduct(ctx context.Context, req models.UpdateProductRequest) error {
	var exists bool
	err := r.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM products WHERE id = $1)`, req.ID).Scan(&exists)
	if err != nil {
		return err
	}
	if !exists {
		return ErrNotFound
	}
	if req.ClassNodeID != 0 {
		var classExists bool
		err = r.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM classifier_nodes WHERE id = $1)`, req.ClassNodeID).Scan(&classExists)
		if err != nil {
			return err
		}
		if !classExists {
			return ErrNotFound
		}
	}

	query := `
		UPDATE products
		SET name = $1, class_node_id = $2, unit_type = $3, weight_per_meter = $4,
		    piece_length = $5, default_unit_id = $6, updated_at = now()
		WHERE id = $7
	`
	_, err = r.db.ExecContext(ctx, query,
		req.Name,
		req.ClassNodeID,
		req.UnitType,
		req.WeightPerMeter,
		req.PieceLength,
		req.DefaultUnitID,
		req.ID,
	)
	return err
}

func (r *PostgresRepository) DeleteProduct(ctx context.Context, id int) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM products WHERE id = $1`, id)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *PostgresRepository) GetProductsByClass(ctx context.Context, classNodeID int) ([]*models.Product, error) {
	query := `
        SELECT id, name, class_node_id, unit_type, weight_per_meter, piece_length, default_unit_id, created_at, updated_at
        FROM products
        WHERE class_node_id = $1
        ORDER BY name
    `
	rows, err := r.db.QueryContext(ctx, query, classNodeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []*models.Product
	for rows.Next() {
		var p models.Product
		if err := rows.Scan(
			&p.ID,
			&p.Name,
			&p.ClassNodeID,
			&p.UnitType,
			&p.WeightPerMeter,
			&p.PieceLength,
			&p.DefaultUnitID,
			&p.CreatedAt,
			&p.UpdatedAt,
		); err != nil {
			return nil, err
		}
		products = append(products, &p)
	}
	return products, rows.Err()
}
