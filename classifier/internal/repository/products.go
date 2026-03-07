package repository

import (
	"classifier/internal/models"
	"context"
	"database/sql"
)

func (r *PostgresRepository) CreateProduct(ctx context.Context, req models.CreateProductRequest) (*models.Product, error) {
	var product models.Product
	query := `
		INSERT INTO products (unit_type, weight_per_meter, piece_length, default_unit_id)
		VALUES ($1, $2, $3, $4)
		RETURNING id, unit_type, weight_per_meter, piece_length, default_unit_id, created_at, updated_at
	`
	err := r.db.QueryRowContext(ctx, query, req.UnitType, req.WeightPerMeter, req.PieceLength, req.DefaultUnitID).Scan(
		&product.ID, &product.UnitType, &product.WeightPerMeter,
		&product.PieceLength, &product.DefaultUnitID, &product.CreatedAt, &product.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &product, nil
}

func (r *PostgresRepository) GetProduct(ctx context.Context, id int) (*models.Product, error) {
	var product models.Product
	query := `SELECT id, unit_type, weight_per_meter, piece_length, default_unit_id, created_at, updated_at FROM products WHERE id = $1`
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&product.ID, &product.UnitType, &product.WeightPerMeter,
		&product.PieceLength, &product.DefaultUnitID, &product.CreatedAt, &product.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &product, nil
}
