package repository

import (
	"classifier/internal/models"
	"context"
	"database/sql"
	"fmt"
)

// TODO: in get ancessors descendors add products
const trashNodeID = 1

func (r *PostgresRepository) CreateNode(ctx context.Context, req models.CreateNodeRequest) (*models.Node, error) {
	var parent *models.Node
	var unitID *int
	needUpdateParent := false

	if req.ParentID != nil {
		var err error
		parent, err = r.getNodeByID(ctx, *req.ParentID)
		if err != nil {
			return nil, err
		}
		if err := r.checkChildCompatibility(parent); err != nil {
			return nil, err
		}
		if parent.IsTerminal == nil {
			needUpdateParent = true
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
		INSERT INTO classifier_nodes (name, parent_id, is_terminal, unit_id, sort_order)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, name, parent_id, is_terminal, unit_id, sort_order, created_at, updated_at
	`

	err := r.db.QueryRowContext(ctx, query,
		req.Name,
		req.ParentID,
		nil,
		unitID,
		sortOrder,
	).Scan(
		&node.ID,
		&node.Name,
		&node.ParentID,
		&node.IsTerminal,
		&node.UnitID,
		&node.SortOrder,
		&node.CreatedAt,
		&node.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if needUpdateParent {
		newParentTerminal := false
		queryToUpdate := `
			UPDATE classifier_nodes SET is_terminal = $1 WHERE id = $2
		`
		_, err = r.db.ExecContext(ctx, queryToUpdate, newParentTerminal, *req.ParentID)
		if err != nil {
			return nil, fmt.Errorf("failed to update parent is_terminal: %w", err)
		}
	}

	return &node, nil

}

func (r *PostgresRepository) GetNode(ctx context.Context, id int) (*models.Node, error) {
	return r.getNodeByID(ctx, id)
}

func (r *PostgresRepository) GetChildren(ctx context.Context, parentID int) ([]*models.Node, error) {
	query := `
		SELECT id, name, parent_id, is_terminal, unit_id, sort_order, created_at, updated_at
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
		var node models.Node
		if err := rows.Scan(
			&node.ID,
			&node.Name,
			&node.ParentID,
			&node.IsTerminal,
			&node.UnitID,
			&node.SortOrder,
			&node.CreatedAt,
			&node.UpdatedAt,
		); err != nil {
			return nil, err
		}
		children = append(children, &node)
	}
	return children, rows.Err()
}

func (r *PostgresRepository) GetAllDescendants(ctx context.Context, id int) ([]*models.Node, error) {
	query := `
		WITH RECURSIVE descendants AS (
			SELECT id, name, parent_id, is_terminal, unit_id, sort_order, created_at, updated_at
			FROM classifier_nodes
			WHERE parent_id = $1
			UNION ALL
			SELECT n.id, n.name, n.parent_id, n.is_terminal, n.unit_id, n.sort_order, n.created_at, n.updated_at
			FROM classifier_nodes n
			INNER JOIN descendants d ON n.parent_id = d.id
		)
		SELECT id, name, parent_id, is_terminal, unit_id, sort_order, created_at, updated_at
		FROM descendants
		ORDER BY sort_order, name
	`
	rows, err := r.db.QueryContext(ctx, query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var nodes []*models.Node
	for rows.Next() {
		var node models.Node
		if err := rows.Scan(
			&node.ID,
			&node.Name,
			&node.ParentID,
			&node.IsTerminal,
			&node.UnitID,
			&node.SortOrder,
			&node.CreatedAt,
			&node.UpdatedAt,
		); err != nil {
			return nil, err
		}
		nodes = append(nodes, &node)
	}
	return nodes, rows.Err()
}

func (r *PostgresRepository) GetAllTerminalDescendants(ctx context.Context, nodeID int) ([]*models.Node, error) {
	query := `
		WITH RECURSIVE descendants AS (
			SELECT id, name, parent_id, is_terminal, unit_id, sort_order, created_at, updated_at
			FROM classifier_nodes
			WHERE parent_id = $1
			UNION ALL
			SELECT n.id, n.name, n.parent_id, n.is_terminal, n.unit_id, n.sort_order, n.created_at, n.updated_at
			FROM classifier_nodes n
			INNER JOIN descendants d ON n.parent_id = d.id
		)
		SELECT id, name, parent_id, is_terminal, unit_id, sort_order, created_at, updated_at
		FROM descendants
		WHERE is_terminal = true
		ORDER BY sort_order, name
	`
	rows, err := r.db.QueryContext(ctx, query, nodeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var nodes []*models.Node
	for rows.Next() {
		var node models.Node
		if err := rows.Scan(
			&node.ID,
			&node.Name,
			&node.ParentID,
			&node.IsTerminal,
			&node.UnitID,
			&node.SortOrder,
			&node.CreatedAt,
			&node.UpdatedAt,
		); err != nil {
			return nil, err
		}
		nodes = append(nodes, &node)
	}
	return nodes, rows.Err()
}

func (r *PostgresRepository) GetAllAncestors(ctx context.Context, id int) ([]*models.Node, error) {
	query := `
		WITH RECURSIVE ancestors AS (
			SELECT id, name, parent_id, is_terminal, unit_id, sort_order, created_at, updated_at
			FROM classifier_nodes
			WHERE id = $1
			UNION ALL
			SELECT n.id, n.name, n.parent_id, n.is_terminal, n.unit_id, n.sort_order, n.created_at, n.updated_at
			FROM classifier_nodes n
			INNER JOIN ancestors a ON n.id = a.parent_id
		)
		SELECT id, name, parent_id, is_terminal, unit_id, sort_order, created_at, updated_at
		FROM ancestors
		WHERE id != $1
		ORDER BY sort_order, name
	`
	rows, err := r.db.QueryContext(ctx, query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var nodes []*models.Node
	for rows.Next() {
		var node models.Node
		if err := rows.Scan(
			&node.ID,
			&node.Name,
			&node.ParentID,
			&node.IsTerminal,
			&node.UnitID,
			&node.SortOrder,
			&node.CreatedAt,
			&node.UpdatedAt,
		); err != nil {
			return nil, err
		}
		nodes = append(nodes, &node)
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
		if err := r.checkChildCompatibility(parent); err != nil {
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

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	selectNodeQuery := `
		SELECT id, parent_id, is_terminal
		FROM classifier_nodes
		WHERE id = $1 FOR UPDATE
	`
	var node models.Node
	err = tx.QueryRowContext(ctx, selectNodeQuery, id).Scan(&node.ID, &node.ParentID, &node.IsTerminal)
	if err == sql.ErrNoRows {
		return ErrNotFound
	}
	if err != nil {
		return err
	}

	parentID := node.ParentID

	if node.IsTerminal != nil && *node.IsTerminal {
		_, err = tx.ExecContext(ctx, `UPDATE products SET class_node_id = $1 WHERE class_node_id = $2`, trashNodeID, id)
		if err != nil {
			return fmt.Errorf("failed to move products to trash: %w", err)
		}
		_, err = tx.ExecContext(ctx, `UPDATE enums SET type_node_id = $1 WHERE type_node_id = $2`, trashNodeID, id)
		if err != nil {
			return fmt.Errorf("failed to move enums to trash: %w", err)
		}
	} else if node.IsTerminal != nil && !*node.IsTerminal {
		selectChildrenQuery := `
			SELECT id FROM classifier_nodes
			WHERE parent_id = $1 FOR UPDATE
		`
		rows, err := tx.QueryContext(ctx, selectChildrenQuery, id)
		if err != nil {
			return err
		}
		var childrenIDs []int
		for rows.Next() {
			var childID int
			if err := rows.Scan(&childID); err != nil {
				rows.Close()
				return err
			}
			childrenIDs = append(childrenIDs, childID)
		}
		rows.Close()
		if err = rows.Err(); err != nil {
			return err
		}

		moveChildQuery := `
			UPDATE classifier_nodes 
			SET parent_id = $1, updated_at = now() 
			WHERE id = $2
		`
		for _, childID := range childrenIDs {
			_, err = tx.ExecContext(ctx, moveChildQuery, trashNodeID, childID)
			if err != nil {
				return fmt.Errorf("failed to move child %d to trash: %w", childID, err)
			}
		}
	}

	deleteNodeQuery := `DELETE FROM classifier_nodes WHERE id = $1`
	result, err := tx.ExecContext(ctx, deleteNodeQuery, id)
	if err != nil {
		return err
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrNotFound
	}

	if parentID != nil {
		countChildrenQuery := `SELECT COUNT(*) FROM classifier_nodes WHERE parent_id = $1`
		var childCount int
		err = tx.QueryRowContext(ctx, countChildrenQuery, *parentID).Scan(&childCount)
		if err != nil {
			return err
		}
		if childCount == 0 {
			resetParentTerminalQuery := `UPDATE classifier_nodes SET is_terminal = NULL WHERE id = $1`
			_, err = tx.ExecContext(ctx, resetParentTerminalQuery, *parentID)
			if err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

func (r *PostgresRepository) UpdateNodeIsTerminal(ctx context.Context, nodeID int, isTerminal *bool) error {
	_, err := r.db.ExecContext(ctx, `UPDATE classifier_nodes SET is_terminal = $1, updated_at = now() WHERE id = $2`, isTerminal, nodeID)
	return err
}

func (r *PostgresRepository) checkChildCompatibility(parent *models.Node) error {
	if parent.ID == trashNodeID {
		return nil
	}
	if parent.IsTerminal != nil && *parent.IsTerminal {
		return ErrInvalidParent
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
