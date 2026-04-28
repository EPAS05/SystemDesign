package repository

import (
	"classifier/internal/models"
	"context"
	"database/sql"
	"fmt"
)

func (r *PostgresRepository) CreateEnum(ctx context.Context, req models.CreateEnumRequest) (*models.Enum, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM classifier_nodes WHERE id = $1)`, req.TypeNodeID).Scan(&exists)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrNotFound
	}

	var enum models.Enum
	query := `
		INSERT INTO enums (name, description, type_node_id)
		VALUES ($1, $2, $3)
		RETURNING id, name, description, type_node_id, created_at, updated_at
	`
	err = r.db.QueryRowContext(ctx, query, req.Name, req.Description, req.TypeNodeID).Scan(
		&enum.ID, &enum.Name, &enum.Description, &enum.TypeNodeID, &enum.CreatedAt, &enum.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &enum, nil
}

func (r *PostgresRepository) GetEnum(ctx context.Context, id int) (*models.Enum, error) {
	var enum models.Enum
	query := `
		SELECT id, name, description, type_node_id, created_at, updated_at
		FROM enums WHERE id = $1
	`
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&enum.ID, &enum.Name, &enum.Description, &enum.TypeNodeID, &enum.CreatedAt, &enum.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &enum, nil
}

func (r *PostgresRepository) GetAllEnums(ctx context.Context) ([]*models.Enum, error) {
	query := `
		SELECT id, name, description, type_node_id, created_at, updated_at
		FROM enums
		ORDER BY name
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var enums []*models.Enum
	for rows.Next() {
		var e models.Enum
		if err := rows.Scan(&e.ID, &e.Name, &e.Description, &e.TypeNodeID, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, err
		}
		enums = append(enums, &e)
	}
	return enums, rows.Err()
}

func (r *PostgresRepository) GetEnumsByTypeNode(ctx context.Context, typeNodeID int) ([]*models.Enum, error) {
	query := `
		SELECT id, name, description, type_node_id, created_at, updated_at
		FROM enums
		WHERE type_node_id = $1
		ORDER BY name
	`
	rows, err := r.db.QueryContext(ctx, query, typeNodeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var enums []*models.Enum
	for rows.Next() {
		var e models.Enum
		if err := rows.Scan(&e.ID, &e.Name, &e.Description, &e.TypeNodeID, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, err
		}
		enums = append(enums, &e)
	}
	return enums, rows.Err()
}

func (r *PostgresRepository) UpdateEnum(ctx context.Context, req models.UpdateEnumRequest) (*models.Enum, error) {
	if req.TypeNodeID != 0 {
		var exists bool
		err := r.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM classifier_nodes WHERE id = $1)`, req.TypeNodeID).Scan(&exists)
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, fmt.Errorf("type node with id %d does not exist", req.TypeNodeID)
		}
	}

	query := `
		UPDATE enums
		SET name = $1, description = $2, type_node_id = $3, updated_at = now()
		WHERE id = $4
		RETURNING id, name, description, type_node_id, created_at, updated_at
	`

	var enum models.Enum
	err := r.db.QueryRowContext(ctx, query, req.Name, req.Description, req.TypeNodeID, req.ID).Scan(
		&enum.ID, &enum.Name, &enum.Description, &enum.TypeNodeID, &enum.CreatedAt, &enum.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &enum, nil
}

func (r *PostgresRepository) DeleteEnum(ctx context.Context, id int) error {
	query := `
		DELETE FROM enums 
		WHERE id = $1
	`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *PostgresRepository) CreateEnumValue(ctx context.Context, req models.CreateEnumValueRequest) (*models.EnumValue, error) {
	sortOrder := 0
	if req.SortOrder != nil {
		sortOrder = *req.SortOrder
	} else {
		maxOrder, err := r.getMaxEnumValueOrder(ctx, req.EnumID)
		if err != nil {
			return nil, err
		}
		sortOrder = maxOrder
	}

	var ev models.EnumValue
	query := `
        INSERT INTO enum_values (enum_id, value, sort_order)
        VALUES ($1, $2, $3)
        RETURNING id, enum_id, value, sort_order, created_at, updated_at
    `
	err := r.db.QueryRowContext(ctx, query, req.EnumID, req.Value, sortOrder).Scan(
		&ev.ID, &ev.EnumID, &ev.Value, &ev.SortOrder, &ev.CreatedAt, &ev.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &ev, nil
}

func (r *PostgresRepository) GetEnumValues(ctx context.Context, enumID int) ([]*models.EnumValue, error) {
	query := `
        SELECT id, enum_id, value, sort_order, created_at, updated_at
        FROM enum_values
        WHERE enum_id = $1
        ORDER BY sort_order, value
    `
	rows, err := r.db.QueryContext(ctx, query, enumID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var values []*models.EnumValue
	for rows.Next() {
		var v models.EnumValue
		if err := rows.Scan(&v.ID, &v.EnumID, &v.Value, &v.SortOrder, &v.CreatedAt, &v.UpdatedAt); err != nil {
			return nil, err
		}
		values = append(values, &v)
	}
	return values, rows.Err()
}

func (r *PostgresRepository) UpdateEnumValue(ctx context.Context, req models.UpdateEnumValueRequest) (*models.EnumValue, error) {
	var enumValue models.EnumValue
	query := `
		UPDATE enum_values 
		SET value = $1, updated_at = now() 
		WHERE id = $2
		RETURNING id, enum_id, value, sort_order, created_at, updated_at
	`
	err := r.db.QueryRowContext(ctx, query, req.Value, req.ID).Scan(
		&enumValue.ID, &enumValue.EnumID, &enumValue.Value, &enumValue.SortOrder, &enumValue.CreatedAt, &enumValue.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &enumValue, nil
}

func (r *PostgresRepository) DeleteEnumValue(ctx context.Context, id int) error {
	query := `
		DELETE FROM enum_values 
		WHERE id = $1
	`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *PostgresRepository) ReorderEnumValues(ctx context.Context, req models.ReorderEnumValuesRequest) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		UPDATE enum_values 
		SET sort_order = $1 
		WHERE id = $2 AND enum_id = $3
	`
	for i, valueID := range req.ValueIDs {
		_, err := tx.ExecContext(ctx, query, i, valueID, req.EnumID)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *PostgresRepository) GetEnumValue(ctx context.Context, enumValueID int) (*models.EnumValue, error) {
	var enumValue models.EnumValue
	query := `
	SELECT id, enum_id, value, sort_order, created_at, updated_at
	FROM enum_values
    WHERE id = $1
	`
	err := r.db.QueryRowContext(ctx, query, enumValueID).Scan(
		&enumValue.ID, &enumValue.EnumID, &enumValue.Value, &enumValue.SortOrder, &enumValue.CreatedAt, &enumValue.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &enumValue, nil

}

func (r *PostgresRepository) getMaxEnumValueOrder(ctx context.Context, enumID int) (int, error) {
	var max sql.NullInt64
	query := `
		SELECT MAX(sort_order) 
		FROM enum_values 
		WHERE enum_id = $1
	`
	err := r.db.QueryRowContext(ctx, query, enumID).Scan(&max)
	if err != nil {
		return 0, err
	}
	if max.Valid {
		return int(max.Int64) + 1, nil
	}
	return 0, nil
}
