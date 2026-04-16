package repository

import (
	"classifier/internal/models"
	"context"
	"database/sql"
)

func (r *PostgresRepository) CreateCustomer(ctx context.Context, req models.CreateCustomerRequest) (*models.Customer, error) {
	var customer models.Customer
	query := `
		INSERT INTO customers (name, tax_id, address)
		VALUES ($1, $2, $3)
		RETURNING id, name, tax_id, address, created_at, updated_at
	`
	err := r.db.QueryRowContext(ctx, query, req.Name, req.TaxID, req.Address).Scan(
		&customer.ID,
		&customer.Name,
		&customer.TaxID,
		&customer.Address,
		&customer.CreatedAt,
		&customer.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &customer, nil
}

func (r *PostgresRepository) GetCustomer(ctx context.Context, id int) (*models.Customer, error) {
	var customer models.Customer
	query := `
		SELECT id, name, tax_id, address, created_at, updated_at
		FROM customers
		WHERE id = $1
	`
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&customer.ID,
		&customer.Name,
		&customer.TaxID,
		&customer.Address,
		&customer.CreatedAt,
		&customer.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return &customer, nil
}

func (r *PostgresRepository) GetAllCustomers(ctx context.Context) ([]*models.Customer, error) {
	query := `
		SELECT id, name, tax_id, address, created_at, updated_at
		FROM customers
		ORDER BY name
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var customers []*models.Customer
	for rows.Next() {
		var customer models.Customer
		if err := rows.Scan(
			&customer.ID,
			&customer.Name,
			&customer.TaxID,
			&customer.Address,
			&customer.CreatedAt,
			&customer.UpdatedAt,
		); err != nil {
			return nil, err
		}
		customers = append(customers, &customer)
	}

	return customers, rows.Err()
}

func (r *PostgresRepository) UpdateCustomer(ctx context.Context, req models.UpdateCustomerRequest) error {
	query := `
		UPDATE customers
		SET name = $1, tax_id = $2, address = $3, updated_at = now()
		WHERE id = $4
	`
	result, err := r.db.ExecContext(ctx, query, req.Name, req.TaxID, req.Address, req.ID)
	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *PostgresRepository) DeleteCustomer(ctx context.Context, id int) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM customers WHERE id = $1`, id)
	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}
