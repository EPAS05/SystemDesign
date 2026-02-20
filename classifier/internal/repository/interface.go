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
	SetParent(ctx context.Context, req models.SetParentRequest) error
	SetName(ctx context.Context, req models.SetNameRequest) error
	DeleteNode(ctx context.Context, id int) error
}
