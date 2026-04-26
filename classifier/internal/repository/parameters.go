package repository

import (
	"classifier/internal/models"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
)

func (r *PostgresRepository) CreateParameterDefinition(ctx context.Context, req models.CreateParameterDefinitionRequest) (*models.ParameterDefinition, error) {
	_, err := r.getNodeByID(ctx, req.ClassNodeID)
	if err != nil {
		return nil, err
	}
	if req.Constraints != nil && req.Constraints.MinValue != nil && req.Constraints.MaxValue != nil {
		if *req.Constraints.MinValue > *req.Constraints.MaxValue {
			return nil, errors.New("minimum value cannot be greater than maximum value")
		}
	}
	sortOrder := 0
	if req.SortOrder != nil {
		sortOrder = *req.SortOrder
	} else {
		var max sql.NullInt64
		querySort := `
			SELECT MAX(sort_order) 
			FROM parameter_definitions 
			WHERE class_node_id = $1
		`
		err = r.db.QueryRowContext(ctx, querySort, req.ClassNodeID).Scan(&max)
		if err != nil {
			return nil, err
		}
		if max.Valid {
			sortOrder = int(max.Int64) + 1
		}
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var param models.ParameterDefinition
	query := `
        INSERT INTO parameter_definitions (class_node_id, name, description, parameter_type, unit_id, enum_id, is_required, sort_order)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
        RETURNING id, class_node_id, name, description, parameter_type, unit_id, enum_id, is_required, sort_order, created_at, updated_at
    `
	err = tx.QueryRowContext(ctx, query,
		req.ClassNodeID, req.Name, req.Description, req.ParameterType, req.UnitID, req.EnumID, req.IsRequired, sortOrder,
	).Scan(
		&param.ID, &param.ClassNodeID, &param.Name, &param.Description, &param.ParameterType,
		&param.UnitID, &param.EnumID, &param.IsRequired, &param.SortOrder,
		&param.CreatedAt, &param.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if req.ParameterType == "number" && req.Constraints != nil {
		queryConstraint := `
            INSERT INTO parameter_constraints (param_def_id, min_value, max_value)
            VALUES ($1, $2, $3)
            RETURNING id, param_def_id, min_value, max_value, created_at, updated_at
        `
		var cons models.ParameterConstraint
		err = tx.QueryRowContext(ctx, queryConstraint, param.ID, req.Constraints.MinValue, req.Constraints.MaxValue).Scan(
			&cons.ID, &cons.ParamDefID, &cons.MinValue, &cons.MaxValue, &cons.CreatedAt, &cons.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return &param, nil
}

func (r *PostgresRepository) GetParameterDefinition(ctx context.Context, id int) (*models.ParameterDefinition, error) {
	var param models.ParameterDefinition
	query := `
        SELECT id, class_node_id, name, description, parameter_type, unit_id, enum_id, is_required, sort_order, created_at, updated_at
        FROM parameter_definitions WHERE id = $1
    `
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&param.ID, &param.ClassNodeID, &param.Name, &param.Description, &param.ParameterType,
		&param.UnitID, &param.EnumID, &param.IsRequired, &param.SortOrder,
		&param.CreatedAt, &param.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &param, nil
}

func (r *PostgresRepository) GetParameterDefinitionsForClass(ctx context.Context, classNodeID int) ([]*models.ParameterDefinition, error) {
	query := `
        WITH RECURSIVE class_tree AS (
            SELECT id, parent_id, 0 AS depth
            FROM classifier_nodes
            WHERE id = $1
            UNION ALL
            SELECT n.id, n.parent_id, ct.depth + 1
            FROM classifier_nodes n
            INNER JOIN class_tree ct ON n.id = ct.parent_id
        )
        SELECT pd.id, pd.class_node_id, pd.name, pd.description, pd.parameter_type,
               pd.unit_id, pd.enum_id, pd.is_required, pd.sort_order,
               pd.created_at, pd.updated_at,
               pc.min_value, pc.max_value
        FROM parameter_definitions pd
        LEFT JOIN parameter_constraints pc ON pd.id = pc.param_def_id
        WHERE pd.class_node_id IN (SELECT id FROM class_tree)
        ORDER BY (SELECT depth FROM class_tree WHERE id = pd.class_node_id) ASC, pd.sort_order
    `
	rows, err := r.db.QueryContext(ctx, query, classNodeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var params []*models.ParameterDefinition
	for rows.Next() {
		var p models.ParameterDefinition
		var minVal, maxVal sql.NullFloat64
		err := rows.Scan(
			&p.ID, &p.ClassNodeID, &p.Name, &p.Description, &p.ParameterType,
			&p.UnitID, &p.EnumID, &p.IsRequired, &p.SortOrder,
			&p.CreatedAt, &p.UpdatedAt,
			&minVal, &maxVal,
		)
		if err != nil {
			return nil, err
		}
		params = append(params, &p)
	}
	return params, rows.Err()
}

func (r *PostgresRepository) UpdateParameterDefinition(ctx context.Context, req models.UpdateParameterDefinitionRequest) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
        UPDATE parameter_definitions
        SET name = $1, description = $2, unit_id = $3, enum_id = $4, is_required = $5, sort_order = $6, updated_at = now()
        WHERE id = $7
    `
	result, err := tx.ExecContext(ctx, query, req.Name, req.Description, req.UnitID, req.EnumID, req.IsRequired, req.SortOrder, req.ID)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}

	if req.Constraints != nil {
		_, err = tx.ExecContext(ctx, `DELETE FROM parameter_constraints WHERE param_def_id = $1`, req.ID)
		if err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO parameter_constraints (param_def_id, min_value, max_value) VALUES ($1, $2, $3)`,
			req.ID, req.Constraints.MinValue, req.Constraints.MaxValue)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *PostgresRepository) DeleteParameterDefinition(ctx context.Context, id int) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM parameter_definitions WHERE id = $1`, id)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *PostgresRepository) SetParameterValue(ctx context.Context, req models.CreateParameterValueRequest) (*models.ParameterValue, error) {
	var product models.Product
	err := r.db.QueryRowContext(ctx, `SELECT id, class_node_id FROM products WHERE id = $1`, req.ProductID).Scan(&product.ID, &product.ClassNodeID)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	param, err := r.GetParameterDefinition(ctx, req.ParamDefID)
	if err != nil {
		return nil, err
	}

	isAncestor, err := r.isAncestor(ctx, param.ClassNodeID, product.ClassNodeID)
	if err != nil {
		return nil, err
	}
	if !isAncestor {
		return nil, errors.New("parameter not applicable to this product")
	}

	if param.ParameterType == "number" && req.ValueNumeric == nil {
		return nil, errors.New("numeric parameter requires numeric value")
	}
	if param.ParameterType == "enum" && req.ValueEnumID == nil {
		return nil, errors.New("enum parameter requires enum value")
	}

	if param.ParameterType == "number" && req.ValueNumeric != nil {
		var cons models.ParameterConstraint
		err = r.db.QueryRowContext(ctx, `SELECT min_value, max_value FROM parameter_constraints WHERE param_def_id = $1`, param.ID).Scan(&cons.MinValue, &cons.MaxValue)
		if err != nil && err != sql.ErrNoRows {
			return nil, err
		}
		if err == nil {
			if cons.MinValue != nil && *req.ValueNumeric < *cons.MinValue {
				return nil, fmt.Errorf("value %.2f is below minimum %.2f", *req.ValueNumeric, *cons.MinValue)
			}
			if cons.MaxValue != nil && *req.ValueNumeric > *cons.MaxValue {
				return nil, fmt.Errorf("value %.2f exceeds maximum %.2f", *req.ValueNumeric, *cons.MaxValue)
			}
		}
	}

	query := `
        INSERT INTO parameter_values (product_id, param_def_id, value_numeric, value_enum_id)
        VALUES ($1, $2, $3, $4)
        ON CONFLICT (product_id, param_def_id) DO UPDATE SET
            value_numeric = EXCLUDED.value_numeric,
            value_enum_id = EXCLUDED.value_enum_id,
            updated_at = now()
        RETURNING id, product_id, param_def_id, value_numeric, value_enum_id, created_at, updated_at
    `

	var pv models.ParameterValue
	err = r.db.QueryRowContext(ctx, query,
		req.ProductID, req.ParamDefID, req.ValueNumeric, req.ValueEnumID,
	).Scan(
		&pv.ID, &pv.ProductID, &pv.ParamDefID, &pv.ValueNumeric, &pv.ValueEnumID,
		&pv.CreatedAt, &pv.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &pv, nil
}

func (r *PostgresRepository) GetParameterValuesForProduct(ctx context.Context, productID int) ([]*models.ParameterValue, error) {
	rows, err := r.db.QueryContext(ctx, `
        SELECT id, product_id, param_def_id, value_numeric, value_enum_id, created_at, updated_at
        FROM parameter_values
        WHERE product_id = $1
    `, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var values []*models.ParameterValue
	for rows.Next() {
		var v models.ParameterValue
		if err := rows.Scan(&v.ID, &v.ProductID, &v.ParamDefID, &v.ValueNumeric, &v.ValueEnumID, &v.CreatedAt, &v.UpdatedAt); err != nil {
			return nil, err
		}
		values = append(values, &v)
	}
	return values, rows.Err()
}

func (r *PostgresRepository) UpdateParameterValue(ctx context.Context, req models.UpdateParameterValueRequest) error {
	result, err := r.db.ExecContext(ctx, `
        UPDATE parameter_values
        SET value_numeric = $1, value_enum_id = $2, updated_at = now()
        WHERE id = $3
    `, req.ValueNumeric, req.ValueEnumID, req.ID)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *PostgresRepository) DeleteParameterValue(ctx context.Context, id int) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM parameter_values WHERE id = $1`, id)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *PostgresRepository) GetParameterConstraints(ctx context.Context, paramDefID int) (*models.ParameterConstraint, error) {
	var cons models.ParameterConstraint
	query := `
		SELECT id, param_def_id, min_value, max_value, created_at, updated_at 
		FROM parameter_constraints
		WHERE param_def_id = $1
	`
	err := r.db.QueryRowContext(ctx, query, paramDefID).Scan(&cons.ID, &cons.ParamDefID, &cons.MinValue, &cons.MaxValue, &cons.CreatedAt, &cons.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &cons, nil
}

func (r *PostgresRepository) isAncestor(ctx context.Context, ancestorID, nodeID int) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `
		WITH RECURSIVE ancestors AS (
			SELECT id FROM classifier_nodes WHERE id = $1
			UNION ALL
			SELECT parent_id FROM classifier_nodes n
			INNER JOIN ancestors a ON n.id = a.id
			WHERE n.parent_id IS NOT NULL
		)
		SELECT COUNT(*) FROM ancestors WHERE id = $2
	`, nodeID, ancestorID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *PostgresRepository) FindProductsByParameters(ctx context.Context, classNodeID int, filters []models.ParameterFilter) ([]*models.Product, error) {
	if classNodeID <= 0 {
		return nil, fmt.Errorf("class node id must be greater than zero")
	}

	if _, err := r.getNodeByID(ctx, classNodeID); err != nil {
		return nil, err
	}

	var queryBuilder strings.Builder
	queryBuilder.WriteString(`
		SELECT DISTINCT p.id, p.name, p.class_node_id, p.unit_type, p.weight_per_meter, p.piece_length, p.default_unit_id, p.created_at, p.updated_at
		FROM products p
	`)

	args := []interface{}{classNodeID}
	nextArgPos := 2

	for i, filter := range filters {
		if filter.ParamDefID <= 0 {
			return nil, fmt.Errorf("invalid param_def_id in filter %d", i)
		}

		paramDef, err := r.GetParameterDefinition(ctx, filter.ParamDefID)
		if err != nil {
			return nil, err
		}

		operator := strings.TrimSpace(filter.Operator)
		if operator == "" {
			return nil, fmt.Errorf("operator is required in filter %d", i)
		}

		alias := "pv" + strconv.Itoa(i+1)
		queryBuilder.WriteString("INNER JOIN parameter_values ")
		queryBuilder.WriteString(alias)
		queryBuilder.WriteString(" ON ")
		queryBuilder.WriteString(alias)
		queryBuilder.WriteString(".product_id = p.id AND ")
		queryBuilder.WriteString(alias)
		queryBuilder.WriteString(".param_def_id = $")
		queryBuilder.WriteString(strconv.Itoa(nextArgPos))
		queryBuilder.WriteString(" ")
		args = append(args, filter.ParamDefID)
		nextArgPos++

		switch operator {
		case "=":
			if paramDef.ParameterType == "enum" {
				enumValueID, ok := parseIntFilterValue(filter.Value)
				if !ok {
					return nil, fmt.Errorf("enum filter %d requires enum value id", i)
				}
				queryBuilder.WriteString("AND ")
				queryBuilder.WriteString(alias)
				queryBuilder.WriteString(".value_enum_id = $")
				queryBuilder.WriteString(strconv.Itoa(nextArgPos))
				queryBuilder.WriteString(" ")
				args = append(args, enumValueID)
				nextArgPos++
				continue
			}

			numericValue, ok := parseFloatFilterValue(filter.Value)
			if !ok {
				return nil, fmt.Errorf("numeric filter %d requires numeric value", i)
			}
			queryBuilder.WriteString("AND ")
			queryBuilder.WriteString(alias)
			queryBuilder.WriteString(".value_numeric = $")
			queryBuilder.WriteString(strconv.Itoa(nextArgPos))
			queryBuilder.WriteString(" ")
			args = append(args, numericValue)
			nextArgPos++
		case "<", ">", "<=", ">=":
			numericValue, ok := parseFloatFilterValue(filter.Value)
			if !ok {
				return nil, fmt.Errorf("operator %q requires numeric value in filter %d", operator, i)
			}
			queryBuilder.WriteString("AND ")
			queryBuilder.WriteString(alias)
			queryBuilder.WriteString(".value_numeric IS NOT NULL AND ")
			queryBuilder.WriteString(alias)
			queryBuilder.WriteString(".value_numeric ")
			queryBuilder.WriteString(operator)
			queryBuilder.WriteString(" $")
			queryBuilder.WriteString(strconv.Itoa(nextArgPos))
			queryBuilder.WriteString(" ")
			args = append(args, numericValue)
			nextArgPos++
		default:
			return nil, fmt.Errorf("unsupported operator %q in filter %d", operator, i)
		}
	}

	queryBuilder.WriteString("WHERE p.class_node_id = $1 ORDER BY p.name")

	rows, err := r.db.QueryContext(ctx, queryBuilder.String(), args...)
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

func parseFloatFilterValue(value interface{}) (float64, bool) {
	switch v := value.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int8:
		return float64(v), true
	case int16:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	case uint:
		return float64(v), true
	case uint8:
		return float64(v), true
	case uint16:
		return float64(v), true
	case uint32:
		return float64(v), true
	case uint64:
		return float64(v), true
	case string:
		number, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return 0, false
		}
		return number, true
	default:
		return 0, false
	}
}

func parseIntFilterValue(value interface{}) (int, bool) {
	switch v := value.(type) {
	case int:
		return v, true
	case int8:
		return int(v), true
	case int16:
		return int(v), true
	case int32:
		return int(v), true
	case int64:
		return int(v), true
	case uint:
		return int(v), true
	case uint8:
		return int(v), true
	case uint16:
		return int(v), true
	case uint32:
		return int(v), true
	case uint64:
		return int(v), true
	case float64:
		if math.Mod(v, 1) != 0 {
			return 0, false
		}
		return int(v), true
	case float32:
		if math.Mod(float64(v), 1) != 0 {
			return 0, false
		}
		return int(v), true
	case json.Number:
		number, err := v.Int64()
		if err != nil {
			return 0, false
		}
		return int(number), true
	case string:
		number, err := strconv.Atoi(v)
		if err != nil {
			return 0, false
		}
		return number, true
	default:
		return 0, false
	}
}
