package repository

import (
	"classifier/internal/models"
	"context"
)

type Repository interface {
	CreateNode(ctx context.Context, req models.CreateNodeRequest) (*models.Node, error)
	GetNode(ctx context.Context, id int) (*models.Node, error)
	GetChildren(ctx context.Context, id int) ([]*models.Node, error)
	GetParent(ctx context.Context, id int) (*models.Node, error)
	GetAllDescendants(ctx context.Context, id int) ([]*models.Node, error)
	GetAllAncestors(ctx context.Context, id int) ([]*models.Node, error)
	SetParent(ctx context.Context, req models.SetParentRequest) error
	SetName(ctx context.Context, req models.SetNameRequest) error
	SetNodeOrder(ctx context.Context, nodeID int, order int) error
	DeleteNode(ctx context.Context, id int) error

	CreateUnit(ctx context.Context, req models.CreateUnitRequest) (*models.Unit, error)
	GetUnit(ctx context.Context, id int) (*models.Unit, error)
	GetAllUnits(ctx context.Context) ([]*models.Unit, error)
	UpdateUnit(ctx context.Context, req models.UpdateUnitRequest) error
	DeleteUnit(ctx context.Context, id int) error
}
