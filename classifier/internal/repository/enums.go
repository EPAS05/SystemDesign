package repository

import (
	"classifier/internal/models"
	"context"
	"database/sql"
	"fmt"
)

func (r *PostgresRepository) CreateEnum(ctx context.Context, req models.CreateEnumRequest) (*models.Enum, error) {
	parentID := 3
	if req.ParentID != nil {
		parentID = *req.ParentID
	} else {
		switch req.Type {
		case "number":
			parentID = 4
		case "string":
			parentID = 5
		case "image":
			parentID = 6
		}
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var enum models.Enum
	queryEnum := `
        INSERT INTO enums (name, description, type)
        VALUES ($1, $2, $3)
        RETURNING id, name, description, type, created_at, updated_at
    `
	err = tx.QueryRowContext(ctx, queryEnum, req.Name, req.Description, req.Type).Scan(
		&enum.ID, &enum.Name, &enum.Description, &enum.Type, &enum.CreatedAt, &enum.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	queryNode := `
        INSERT INTO classifier_nodes (name, parent_id, node_type, is_terminal, unit_id, sort_order,
                                       object_type, object_id)
        VALUES ($1, $2, 'enum', NULL, NULL,
                COALESCE((SELECT MAX(sort_order)+1 FROM classifier_nodes WHERE parent_id = $2), 0),
                'enum', $3)
        RETURNING id
    `
	var nodeID int
	err = tx.QueryRowContext(ctx, queryNode, req.Name, parentID, enum.ID).Scan(&nodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to create enum node: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return &enum, nil
}

func (r *PostgresRepository) GetEnum(ctx context.Context, id int) (*models.Enum, error) {
	var enum models.Enum
	query := `
        SELECT id, name, description, type, created_at, updated_at
        FROM enums WHERE id = $1
    `
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&enum.ID, &enum.Name, &enum.Description, &enum.Type, &enum.CreatedAt, &enum.UpdatedAt,
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
        SELECT id, name, description, type, created_at, updated_at
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
		if err := rows.Scan(&e.ID, &e.Name, &e.Description, &e.Type, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, err
		}
		enums = append(enums, &e)
	}
	return enums, rows.Err()
}

func (r *PostgresRepository) UpdateEnum(ctx context.Context, req models.UpdateEnumRequest) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	queryEnum := `
        UPDATE enums
        SET name = $1, description = $2, type = $3, updated_at = now()
        WHERE id = $4
    `
	result, err := tx.ExecContext(ctx, queryEnum, req.Name, req.Description, req.Type, req.ID)
	if err != nil {
		return err
	}
	rowsAff, _ := result.RowsAffected()
	if rowsAff == 0 {
		return ErrNotFound
	}

	queryNode := `
        UPDATE classifier_nodes
        SET name = $1, updated_at = now()
        WHERE enum_id = $2
    `
	_, err = tx.ExecContext(ctx, queryNode, req.Name, req.ID)
	if err != nil {
		return err
	}

	return tx.Commit()
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

func (r *PostgresRepository) UpdateEnumValue(ctx context.Context, req models.UpdateEnumValueRequest) error {
	query := `
		UPDATE enum_values 
		SET value = $1, updated_at = now() 
		WHERE id = $2
	`
	result, err := r.db.ExecContext(ctx, query, req.Value, req.ID)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
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
