package repository

import (
	"classifier/internal/models"
	"context"
)

type NodeRepository interface {
	CreateNode(ctx context.Context, req models.CreateNodeRequest) (*models.Node, error)
	GetNode(ctx context.Context, id int) (*models.Node, error)
	GetChildren(ctx context.Context, id int) ([]*models.Node, error)
	GetParent(ctx context.Context, id int) (*models.Node, error)
	GetAllDescendants(ctx context.Context, id int) ([]*models.Node, error)
	GetAllTerminalDescendants(ctx context.Context, nodeID int) ([]*models.Node, error)
	GetAllAncestors(ctx context.Context, id int) ([]*models.Node, error)
	SetParent(ctx context.Context, req models.SetParentRequest) (*models.Node, error)
	SetName(ctx context.Context, req models.SetNameRequest) (*models.Node, error)
	SetNodeOrder(ctx context.Context, nodeID int, order int) error
	DeleteNode(ctx context.Context, id int) error
	UpdateNodeIsTerminal(ctx context.Context, nodeID int, isTerminal *bool) error
}

type UnitRepository interface {
	CreateUnit(ctx context.Context, req models.CreateUnitRequest) (*models.Unit, error)
	GetUnit(ctx context.Context, id int) (*models.Unit, error)
	GetAllUnits(ctx context.Context) ([]*models.Unit, error)
	UpdateUnit(ctx context.Context, req models.UpdateUnitRequest) (*models.Unit, error)
	SetUnit(ctx context.Context, req models.SetUnitRequest) (*models.Node, error)
	SetDefaultUnit(ctx context.Context, req models.SetDefaultUnitRequest) (*models.Product, error)
	DeleteUnit(ctx context.Context, id int) error
}

type EnumRepository interface {
	CreateEnum(ctx context.Context, req models.CreateEnumRequest) (*models.Enum, error)
	GetEnum(ctx context.Context, id int) (*models.Enum, error)
	GetAllEnums(ctx context.Context) ([]*models.Enum, error)
	GetEnumsByTypeNode(ctx context.Context, typeNodeID int) ([]*models.Enum, error)
	UpdateEnum(ctx context.Context, req models.UpdateEnumRequest) (*models.Enum, error)
	DeleteEnum(ctx context.Context, id int) error
	CreateEnumValue(ctx context.Context, req models.CreateEnumValueRequest) (*models.EnumValue, error)
	GetEnumValues(ctx context.Context, enumID int) ([]*models.EnumValue, error)
	GetEnumValue(ctx context.Context, enumValueID int) (*models.EnumValue, error)
	UpdateEnumValue(ctx context.Context, req models.UpdateEnumValueRequest) (*models.EnumValue, error)
	DeleteEnumValue(ctx context.Context, id int) error
	ReorderEnumValues(ctx context.Context, req models.ReorderEnumValuesRequest) error
}

type ProductRepository interface {
	CreateProduct(ctx context.Context, req models.CreateProductRequest) (*models.Product, error)
	GetProduct(ctx context.Context, id int) (*models.Product, error)
	UpdateProduct(ctx context.Context, req models.UpdateProductRequest) (*models.Product, error)
	DeleteProduct(ctx context.Context, id int) error
	GetProductsByClass(ctx context.Context, classNodeID int) ([]*models.Product, error)
}

type ParameterRepository interface {
	CreateParameterDefinition(ctx context.Context, req models.CreateParameterDefinitionRequest) (*models.ParameterDefinition, error)
	GetParameterDefinition(ctx context.Context, id int) (*models.ParameterDefinition, error)
	GetParameterDefinitionsForClass(ctx context.Context, classNodeID int) ([]*models.ParameterDefinition, error)
	UpdateParameterDefinition(ctx context.Context, req models.UpdateParameterDefinitionRequest) (*models.ParameterDefinition, error)
	DeleteParameterDefinition(ctx context.Context, id int) error
	SetParameterValue(ctx context.Context, req models.CreateParameterValueRequest) (*models.ParameterValue, error)
	GetParameterValuesForProduct(ctx context.Context, productNodeID int) ([]*models.ParameterValue, error)
	UpdateParameterValue(ctx context.Context, req models.UpdateParameterValueRequest) (*models.ParameterValue, error)
	DeleteParameterValue(ctx context.Context, id int) error
	GetParameterConstraints(ctx context.Context, paramDefID int) (*models.ParameterConstraint, error)
	FindProductsByParameters(ctx context.Context, classNodeID int, filters []models.ParameterFilter) ([]*models.Product, error)
}

type CustomerRepository interface {
	CreateCustomer(ctx context.Context, req models.CreateCustomerRequest) (*models.Customer, error)
	GetCustomer(ctx context.Context, id int) (*models.Customer, error)
	GetAllCustomers(ctx context.Context) ([]*models.Customer, error)
	UpdateCustomer(ctx context.Context, req models.UpdateCustomerRequest) (*models.Customer, error)
	DeleteCustomer(ctx context.Context, id int) error
}

type InvoiceRepository interface {
	CreateInvoice(ctx context.Context, req models.CreateInvoiceRequest) (*models.Invoice, error)
	GetInvoice(ctx context.Context, id int) (*models.Invoice, error)
	GetAllInvoices(ctx context.Context) ([]*models.Invoice, error)
	UpdateInvoice(ctx context.Context, req models.UpdateInvoiceRequest) error
	DeleteInvoice(ctx context.Context, id int) error
	AddInvoiceItem(ctx context.Context, req models.CreateInvoiceItemRequest) (*models.InvoiceItem, error)
	GetInvoiceItems(ctx context.Context, invoiceID int) ([]*models.InvoiceItem, error)
	UpdateInvoiceItem(ctx context.Context, req models.UpdateInvoiceItemRequest) error
	DeleteInvoiceItem(ctx context.Context, id int) error
	RecalculateInvoiceTotal(ctx context.Context, invoiceID int) error
}

type Repository interface {
	NodeRepository
	UnitRepository
	EnumRepository
	ProductRepository
	ParameterRepository
	CustomerRepository
	InvoiceRepository
}
