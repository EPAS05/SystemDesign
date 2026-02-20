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

	if req.ParentID != nil {
		parent, err := r.getNodeByID(ctx, *req.ParentID)
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

	var node models.Node
	//TODO
	query := `
    
    `
	err := r.db.QueryRowContext(
		ctx, query,
		req.Name,
		req.ParentID,
		req.NodeType,
		req.IsTerminal,
	).Scan(
		&node.ID,
		&node.Name,
		&node.ParentID,
		&node.NodeType,
		&node.IsTerminal,
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

func (r *PostgresRepository) GetChildren(ctx context.Context, id int) ([]*models.Node, error) {
	//TODO
	query := `
	
	`
	rows, err := r.db.QueryContext(ctx, query, id)
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
			&node.NodeType,
			&node.IsTerminal,
			&node.CreatedAt,
			&node.UpdatedAt,
		); err != nil {
			return nil, err
		}
		children = append(children, &node)
	}
	return children, rows.Err()
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
	//TODO
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
	//TODO
	query := `

	`
	_, err = r.db.ExecContext(ctx, query, req.NewParentID, req.NodeId)
	return err

}

func (r *PostgresRepository) SetName(ctx context.Context, req models.SetNameRequest) error {
	//TODO
	query := `
	
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

func (r *PostgresRepository) DeleteNode(ctx context.Context, id int) error {
	//TODO
	if id == trashNodeID {
		return ErrCannotDeleteTrash
	}

	node, err := r.getNodeByID(ctx, id)
	if err != nil {
		return err
	}

	//TODO
	queryToMove := `
	
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

	//TODO
	queryToDelete := `
	
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

func (r *PostgresRepository) getNodeByID(ctx context.Context, id int) (*models.Node, error) {
	query := `

	`
	//TODO

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
	//TODO
	query := `
	
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
