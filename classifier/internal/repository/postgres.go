package repository

import (
	"classifier/internal/models"
	"context"
	"database/sql"
	"errors"
)

var (
	ErrNotFound          = errors.New("node not found")
	ErrInvalidParent     = errors.New("invalid parent node")
	ErrTypeMismatch      = errors.New("child type not allowed for this parent")
	ErrCannotDeleteTrash = errors.New("cannot delete trash node")
	ErrCycleDetected     = errors.New("Cycle detected")
)

const trashNodeID = 1

type PostgresRepository struct {
	db *sql.DB
}

var _ Repository = (*PostgresRepository)(nil)

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) CreateNode(ctx context.Context, req models.CreateNodeRequest) (*models.Node, error) {
	var parent *models.Node
	var unitID *int

	if req.ParentID != nil {
		var err error
		parent, err = r.getNodeByID(ctx, *req.ParentID)
		if err != nil {
			return nil, err
		}
		if parent.NodeType != models.TypeMetaclass {
			return nil, ErrInvalidParent
		}
		if err := r.checkChildCompatibility(parent, req.NodeType); err != nil {
			return nil, err
		}
	}

	if req.UnitID != nil {
		unitID = req.UnitID
	} else if parent != nil {
		unitID = parent.UnitID
	}

	sortOrder := 0
	if req.SortOrder != nil {
		sortOrder = *req.SortOrder
	} else if parent != nil {
		maxOrder, err := r.getMaxSortOrder(ctx, *req.ParentID)
		if err != nil {
			return nil, err
		}
		sortOrder = maxOrder
	}

	var node models.Node

	query := `
		INSERT INTO classifier_nodes (name, parent_id, node_type, is_terminal, unit_id, sort_order)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, name, parent_id, node_type, is_terminal, unit_id, sort_order, created_at, updated_at
	`
	err := r.db.QueryRowContext(ctx, query,
		req.Name,
		req.ParentID,
		req.NodeType,
		req.IsTerminal,
		unitID,
		sortOrder,
	).Scan(
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
	if err != nil {
		return nil, err
	}
	return &node, nil
}

func (r *PostgresRepository) GetNode(ctx context.Context, id int) (*models.Node, error) {
	return r.getNodeByID(ctx, id)
}

func (r *PostgresRepository) GetChildren(ctx context.Context, parentID int) ([]*models.Node, error) {
	query := `
		SELECT id, name, parent_id, node_type, is_terminal, unit_id, sort_order, created_at, updated_at
		FROM classifier_nodes
		WHERE parent_id = $1
		ORDER BY sort_order, name
	`
	rows, err := r.db.QueryContext(ctx, query, parentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var children []*models.Node
	for rows.Next() {
		var n models.Node
		if err := rows.Scan(
			&n.ID,
			&n.Name,
			&n.ParentID,
			&n.NodeType,
			&n.IsTerminal,
			&n.UnitID,
			&n.SortOrder,
			&n.CreatedAt,
			&n.UpdatedAt,
		); err != nil {
			return nil, err
		}
		children = append(children, &n)
	}
	return children, rows.Err()
}

func (r *PostgresRepository) GetAllDescendants(ctx context.Context, id int) ([]*models.Node, error) {
	query := `
		WITH RECURSIVE descendants AS (
			SELECT id, name, parent_id, node_type, is_terminal, unit_id, sort_order, created_at, updated_at
			FROM classifier_nodes
			WHERE parent_id = $1
			UNION ALL
			SELECT n.id, n.name, n.parent_id, n.node_type, n.is_terminal, n.unit_id, n.sort_order, n.created_at, n.updated_at
			FROM classifier_nodes n
			INNER JOIN descendants d ON n.parent_id = d.id
		)
		SELECT id, name, parent_id, node_type, is_terminal, unit_id, sort_order, created_at, updated_at
		FROM descendants
	`
	rows, err := r.db.QueryContext(ctx, query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var nodes []*models.Node
	for rows.Next() {
		var n models.Node
		if err := rows.Scan(
			&n.ID,
			&n.Name,
			&n.ParentID,
			&n.NodeType,
			&n.IsTerminal,
			&n.UnitID,
			&n.SortOrder,
			&n.CreatedAt,
			&n.UpdatedAt,
		); err != nil {
			return nil, err
		}
		nodes = append(nodes, &n)
	}
	return nodes, rows.Err()
}

func (r *PostgresRepository) GetAllAncestors(ctx context.Context, id int) ([]*models.Node, error) {
	query := `
		WITH RECURSIVE ancestors AS (
			SELECT id, name, parent_id, node_type, is_terminal, unit_id, sort_order, created_at, updated_at
			FROM classifier_nodes
			WHERE id = $1
			UNION ALL
			SELECT n.id, n.name, n.parent_id, n.node_type, n.is_terminal, n.unit_id, n.sort_order, n.created_at, n.updated_at
			FROM classifier_nodes n
			INNER JOIN ancestors a ON n.id = a.parent_id
		)
		SELECT id, name, parent_id, node_type, is_terminal, unit_id, sort_order, created_at, updated_at
		FROM ancestors
		WHERE id != $1
	`
	rows, err := r.db.QueryContext(ctx, query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var nodes []*models.Node
	for rows.Next() {
		var n models.Node
		if err := rows.Scan(
			&n.ID,
			&n.Name,
			&n.ParentID,
			&n.NodeType,
			&n.IsTerminal,
			&n.UnitID,
			&n.SortOrder,
			&n.CreatedAt,
			&n.UpdatedAt,
		); err != nil {
			return nil, err
		}
		nodes = append(nodes, &n)
	}
	return nodes, rows.Err()
}

func (r *PostgresRepository) GetParent(ctx context.Context, id int) (*models.Node, error) {
	node, err := r.getNodeByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if node.ParentID == nil {
		return nil, nil
	}
	return r.getNodeByID(ctx, *node.ParentID)
}

func (r *PostgresRepository) SetParent(ctx context.Context, req models.SetParentRequest) error {
	node, err := r.getNodeByID(ctx, req.NodeId)
	if err != nil {
		return err
	}

	if (node.ParentID == nil && req.NewParentID == nil) ||
		(node.ParentID != nil && req.NewParentID != nil && *node.ParentID == *req.NewParentID) {
		return nil
	}

	if req.NewParentID != nil {
		parent, err := r.getNodeByID(ctx, *req.NewParentID)
		if err != nil {
			return ErrNotFound
		}
		if parent.NodeType != models.TypeMetaclass {
			return ErrInvalidParent
		}
		if err := r.checkChildCompatibility(parent, node.NodeType); err != nil {
			return err
		}
		if err := r.checkCycle(ctx, node.ID, *req.NewParentID); err != nil {
			return err
		}
	}

	query := `
		UPDATE classifier_nodes 
		SET parent_id = $1, updated_at = now() 
		WHERE id = $2;
	`
	_, err = r.db.ExecContext(ctx, query, req.NewParentID, req.NodeId)
	return err

}

func (r *PostgresRepository) SetName(ctx context.Context, req models.SetNameRequest) error {
	query := `
		UPDATE classifier_nodes 
		SET name = $1, updated_at = now() 
		WHERE id = $2;
	`
	result, err := r.db.ExecContext(ctx, query, req.Name, req.NodeId)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *PostgresRepository) SetNodeOrder(ctx context.Context, nodeID int, order int) error {
	query := `
		UPDATE classifier_nodes
		SET sort_order = $1, updated_at = now()
		WHERE id = $2
	`
	result, err := r.db.ExecContext(ctx, query, order, nodeID)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *PostgresRepository) DeleteNode(ctx context.Context, id int) error {
	if id == trashNodeID {
		return ErrCannotDeleteTrash
	}

	node, err := r.getNodeByID(ctx, id)
	if err != nil {
		return err
	}

	queryToMove := `
		UPDATE classifier_nodes 
		SET parent_id = $1, updated_at = now() 
		WHERE id = $2;
	`

	if node.NodeType == models.TypeMetaclass {
		children, err := r.GetChildren(ctx, id)
		if err != nil {
			return err
		}
		if len(children) > 0 {
			for _, child := range children {
				_, err := r.db.ExecContext(ctx, queryToMove, trashNodeID, child.ID)
				if err != nil {
					return errors.New("failed to move child to trash")
				}
			}
		}
	}

	queryToDelete := `
		DELETE FROM classifier_nodes 
		WHERE id = $1;
	`
	result, err := r.db.ExecContext(ctx, queryToDelete, id)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

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
		ORDER BY name;
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

func (r *PostgresRepository) UpdateUnit(ctx context.Context, req models.UpdateUnitRequest) error {
	query := `
		UPDATE units 
		SET name = $1, multiplier = $2, updated_at = now() 
		WHERE id = $3;
	`
	result, err := r.db.ExecContext(ctx, query, req.Name, req.Multiplier, req.ID)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
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

func (r *PostgresRepository) getNodeByID(ctx context.Context, id int) (*models.Node, error) {
	query := `
		SELECT id, name, parent_id, node_type, is_terminal, created_at, updated_at
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

func (r *PostgresRepository) checkChildCompatibility(parent *models.Node, childType models.NodeType) error {
	if parent.ID == trashNodeID {
		return nil
	}

	if parent.IsTerminal == nil {
		return errors.New("error getting parent")
	}

	if *parent.IsTerminal {
		if childType != models.TypeLeaf {
			return ErrTypeMismatch
		}
	} else {
		if childType != models.TypeMetaclass {
			return ErrTypeMismatch
		}
	}
	return nil
}

func (r *PostgresRepository) checkCycle(ctx context.Context, nodeID, newParentID int) error {
	query := `
		SELECT parent_id 
		FROM classifier_nodes 
		WHERE id = $1;
	`
	currentID := newParentID
	for currentID != 0 {
		if currentID == nodeID {
			return ErrCycleDetected
		}
		var parentID *int
		err := r.db.QueryRowContext(ctx, query, currentID).Scan(&parentID)
		if err != nil {
			if err == sql.ErrNoRows {
				break
			}
			return err
		}
		if parentID == nil {
			break
		}
		currentID = *parentID
	}
	return nil
}

func (r *PostgresRepository) getMaxSortOrder(ctx context.Context, parentID int) (int, error) {
	var max sql.NullInt64
	query := `
		SELECT MAX(sort_order) 
		FROM classifier_nodes 
		WHERE parent_id = $1;
	`
	err := r.db.QueryRowContext(ctx, query, parentID).Scan(&max)
	if err != nil {
		return 0, err
	}
	if max.Valid {
		return int(max.Int64) + 1, nil
	}
	return 0, nil
}
