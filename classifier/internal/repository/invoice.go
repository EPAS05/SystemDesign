package repository

import (
	"classifier/internal/models"
	"context"
	"database/sql"
)

func (r *PostgresRepository) CreateInvoice(ctx context.Context, req models.CreateInvoiceRequest) (*models.Invoice, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM customers WHERE id = $1)`, req.CustomerID).Scan(&exists)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrCustomerNotFound
	}

	var invoice models.Invoice
	query := `
		INSERT INTO invoices (
			invoice_number, invoice_date, invoice_type, status, customer_id, currency,
			discount_total, tax_rate, comment
		)
		VALUES ($1, $2, $3, 'draft', $4, $5, $6, $7, $8)
		RETURNING id, invoice_number, invoice_date, invoice_type, status, customer_id,
			currency, total_amount, discount_total, tax_rate, tax_amount, comment, created_at, updated_at
	`
	err = r.db.QueryRowContext(ctx, query,
		req.InvoiceNumber,
		req.InvoiceDate,
		req.InvoiceType,
		req.CustomerID,
		req.Currency,
		req.DiscountTotal,
		req.TaxRate,
		req.Comment,
	).Scan(
		&invoice.ID,
		&invoice.InvoiceNumber,
		&invoice.InvoiceDate,
		&invoice.InvoiceType,
		&invoice.Status,
		&invoice.CustomerID,
		&invoice.Currency,
		&invoice.TotalAmount,
		&invoice.DiscountTotal,
		&invoice.TaxRate,
		&invoice.TaxAmount,
		&invoice.Comment,
		&invoice.CreatedAt,
		&invoice.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &invoice, nil
}

func (r *PostgresRepository) GetInvoice(ctx context.Context, id int) (*models.Invoice, error) {
	var invoice models.Invoice
	query := `
		SELECT id, invoice_number, invoice_date, invoice_type, status, customer_id,
			currency, total_amount, discount_total, tax_rate, tax_amount, comment, created_at, updated_at
		FROM invoices
		WHERE id = $1
	`
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&invoice.ID,
		&invoice.InvoiceNumber,
		&invoice.InvoiceDate,
		&invoice.InvoiceType,
		&invoice.Status,
		&invoice.CustomerID,
		&invoice.Currency,
		&invoice.TotalAmount,
		&invoice.DiscountTotal,
		&invoice.TaxRate,
		&invoice.TaxAmount,
		&invoice.Comment,
		&invoice.CreatedAt,
		&invoice.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return &invoice, nil
}

func (r *PostgresRepository) GetAllInvoices(ctx context.Context) ([]*models.Invoice, error) {
	query := `
		SELECT id, invoice_number, invoice_date, invoice_type, status, customer_id,
			currency, total_amount, discount_total, tax_rate, tax_amount, comment, created_at, updated_at
		FROM invoices
		ORDER BY invoice_date DESC, id DESC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var invoices []*models.Invoice
	for rows.Next() {
		var invoice models.Invoice
		if err := rows.Scan(
			&invoice.ID,
			&invoice.InvoiceNumber,
			&invoice.InvoiceDate,
			&invoice.InvoiceType,
			&invoice.Status,
			&invoice.CustomerID,
			&invoice.Currency,
			&invoice.TotalAmount,
			&invoice.DiscountTotal,
			&invoice.TaxRate,
			&invoice.TaxAmount,
			&invoice.Comment,
			&invoice.CreatedAt,
			&invoice.UpdatedAt,
		); err != nil {
			return nil, err
		}
		invoices = append(invoices, &invoice)
	}

	return invoices, rows.Err()
}

func (r *PostgresRepository) UpdateInvoice(ctx context.Context, req models.UpdateInvoiceRequest) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var exists bool
	err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM customers WHERE id = $1)`, req.CustomerID).Scan(&exists)
	if err != nil {
		return err
	}
	if !exists {
		return ErrCustomerNotFound
	}

	query := `
		UPDATE invoices
		SET invoice_number = $1,
			invoice_date = $2,
			invoice_type = $3,
			status = $4,
			customer_id = $5,
			currency = $6,
			discount_total = $7,
			tax_rate = $8,
			comment = $9,
			updated_at = now()
		WHERE id = $10
	`
	result, err := tx.ExecContext(ctx, query,
		req.InvoiceNumber,
		req.InvoiceDate,
		req.InvoiceType,
		req.Status,
		req.CustomerID,
		req.Currency,
		req.DiscountTotal,
		req.TaxRate,
		req.Comment,
		req.ID,
	)
	if err != nil {
		return err
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrNotFound
	}

	if err := recalculateInvoiceTotalTx(ctx, tx, req.ID); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *PostgresRepository) DeleteInvoice(ctx context.Context, id int) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM invoices WHERE id = $1`, id)
	if err != nil {
		return err
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *PostgresRepository) AddInvoiceItem(ctx context.Context, req models.CreateInvoiceItemRequest) (*models.InvoiceItem, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var status string
	err = tx.QueryRowContext(ctx, `SELECT status FROM invoices WHERE id = $1 FOR UPDATE`, req.InvoiceID).Scan(&status)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if status != "draft" {
		return nil, ErrInvoiceNotDraft
	}

	var productExists bool
	err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM products WHERE id = $1)`, req.ProductID).Scan(&productExists)
	if err != nil {
		return nil, err
	}
	if !productExists {
		return nil, ErrNotFound
	}

	discountPercent := 0.0
	if req.DiscountPercent != nil {
		discountPercent = *req.DiscountPercent
	}
	totalLine := req.Quantity * req.UnitPrice * (1 - discountPercent/100)

	var item models.InvoiceItem
	query := `
		INSERT INTO invoice_items (invoice_id, product_id, quantity, unit_price, discount_percent, total_line)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, invoice_id, product_id, quantity, unit_price, discount_percent, total_line, created_at, updated_at
	`
	err = tx.QueryRowContext(ctx, query,
		req.InvoiceID,
		req.ProductID,
		req.Quantity,
		req.UnitPrice,
		req.DiscountPercent,
		totalLine,
	).Scan(
		&item.ID,
		&item.InvoiceID,
		&item.ProductID,
		&item.Quantity,
		&item.UnitPrice,
		&item.DiscountPercent,
		&item.TotalLine,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if err := recalculateInvoiceTotalTx(ctx, tx, req.InvoiceID); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &item, nil
}

func (r *PostgresRepository) GetInvoiceItems(ctx context.Context, invoiceID int) ([]*models.InvoiceItem, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM invoices WHERE id = $1)`, invoiceID).Scan(&exists)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrNotFound
	}

	query := `
		SELECT id, invoice_id, product_id, quantity, unit_price, discount_percent, total_line, created_at, updated_at
		FROM invoice_items
		WHERE invoice_id = $1
		ORDER BY id
	`
	rows, err := r.db.QueryContext(ctx, query, invoiceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*models.InvoiceItem
	for rows.Next() {
		var item models.InvoiceItem
		if err := rows.Scan(
			&item.ID,
			&item.InvoiceID,
			&item.ProductID,
			&item.Quantity,
			&item.UnitPrice,
			&item.DiscountPercent,
			&item.TotalLine,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, &item)
	}

	return items, rows.Err()
}

func (r *PostgresRepository) UpdateInvoiceItem(ctx context.Context, req models.UpdateInvoiceItemRequest) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var invoiceID int
	var status string
	err = tx.QueryRowContext(ctx, `
		SELECT ii.invoice_id, i.status
		FROM invoice_items ii
		JOIN invoices i ON i.id = ii.invoice_id
		WHERE ii.id = $1
		FOR UPDATE OF i
	`, req.ID).Scan(&invoiceID, &status)
	if err == sql.ErrNoRows {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	if status != "draft" {
		return ErrInvoiceNotDraft
	}

	discountPercent := 0.0
	if req.DiscountPercent != nil {
		discountPercent = *req.DiscountPercent
	}
	totalLine := req.Quantity * req.UnitPrice * (1 - discountPercent/100)

	query := `
		UPDATE invoice_items
		SET quantity = $1,
			unit_price = $2,
			discount_percent = $3,
			total_line = $4,
			updated_at = now()
		WHERE id = $5
	`
	result, err := tx.ExecContext(ctx, query,
		req.Quantity,
		req.UnitPrice,
		req.DiscountPercent,
		totalLine,
		req.ID,
	)
	if err != nil {
		return err
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrNotFound
	}

	if err := recalculateInvoiceTotalTx(ctx, tx, invoiceID); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *PostgresRepository) DeleteInvoiceItem(ctx context.Context, id int) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var invoiceID int
	var status string
	err = tx.QueryRowContext(ctx, `
		SELECT ii.invoice_id, i.status
		FROM invoice_items ii
		JOIN invoices i ON i.id = ii.invoice_id
		WHERE ii.id = $1
		FOR UPDATE OF i
	`, id).Scan(&invoiceID, &status)
	if err == sql.ErrNoRows {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	if status != "draft" {
		return ErrInvoiceNotDraft
	}

	result, err := tx.ExecContext(ctx, `DELETE FROM invoice_items WHERE id = $1`, id)
	if err != nil {
		return err
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrNotFound
	}

	if err := recalculateInvoiceTotalTx(ctx, tx, invoiceID); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *PostgresRepository) RecalculateInvoiceTotal(ctx context.Context, invoiceID int) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := recalculateInvoiceTotalTx(ctx, tx, invoiceID); err != nil {
		return err
	}

	return tx.Commit()
}

func recalculateInvoiceTotalTx(ctx context.Context, tx *sql.Tx, invoiceID int) error {
	var discountTotal sql.NullFloat64
	var taxRate sql.NullFloat64

	err := tx.QueryRowContext(ctx, `SELECT discount_total, tax_rate FROM invoices WHERE id = $1 FOR UPDATE`, invoiceID).Scan(&discountTotal, &taxRate)
	if err == sql.ErrNoRows {
		return ErrNotFound
	}
	if err != nil {
		return err
	}

	var itemsSum sql.NullFloat64
	err = tx.QueryRowContext(ctx, `SELECT COALESCE(SUM(total_line), 0) FROM invoice_items WHERE invoice_id = $1`, invoiceID).Scan(&itemsSum)
	if err != nil {
		return err
	}

	subtotal := 0.0
	if itemsSum.Valid {
		subtotal = itemsSum.Float64
	}

	discount := 0.0
	if discountTotal.Valid {
		discount = discountTotal.Float64
	}

	rate := 0.0
	if taxRate.Valid {
		rate = taxRate.Float64
	}

	taxableAmount := subtotal - discount
	if taxableAmount < 0 {
		taxableAmount = 0
	}
	taxAmount := taxableAmount * rate / 100
	totalAmount := taxableAmount + taxAmount

	result, err := tx.ExecContext(ctx, `
		UPDATE invoices
		SET tax_amount = $1,
			total_amount = $2,
			updated_at = now()
		WHERE id = $3
	`, taxAmount, totalAmount, invoiceID)
	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}
