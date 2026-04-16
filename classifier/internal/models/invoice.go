package models

import "time"

type Customer struct {
	ID        int       `db:"id"`
	Name      string    `db:"name"`
	TaxID     *string   `db:"tax_id"`
	Address   *string   `db:"address"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type CreateCustomerRequest struct {
	Name    string
	TaxID   *string
	Address *string
}

type UpdateCustomerRequest struct {
	ID      int
	Name    string
	TaxID   *string
	Address *string
}

type Invoice struct {
	ID            int       `db:"id"`
	InvoiceNumber string    `db:"invoice_number"`
	InvoiceDate   time.Time `db:"invoice_date"`
	InvoiceType   string    `db:"invoice_type"` // incoming, outgoing, return
	Status        string    `db:"status"`       // draft, confirmed, paid, shipped, cancelled
	CustomerID    int       `db:"customer_id"`
	Currency      string    `db:"currency"`
	TotalAmount   float64   `db:"total_amount"`
	DiscountTotal *float64  `db:"discount_total"`
	TaxRate       *float64  `db:"tax_rate"`
	TaxAmount     *float64  `db:"tax_amount"`
	Comment       *string   `db:"comment"`
	CreatedAt     time.Time `db:"created_at"`
	UpdatedAt     time.Time `db:"updated_at"`
}

type CreateInvoiceRequest struct {
	InvoiceNumber string
	InvoiceDate   time.Time
	InvoiceType   string
	CustomerID    int
	Currency      string
	DiscountTotal *float64
	TaxRate       *float64
	Comment       *string
}

type UpdateInvoiceRequest struct {
	ID            int
	InvoiceNumber string
	InvoiceDate   time.Time
	InvoiceType   string
	Status        string
	CustomerID    int
	Currency      string
	DiscountTotal *float64
	TaxRate       *float64
	Comment       *string
}

type InvoiceItem struct {
	ID              int       `db:"id"`
	InvoiceID       int       `db:"invoice_id"`
	ProductID       int       `db:"product_id"`
	Quantity        float64   `db:"quantity"`
	UnitPrice       float64   `db:"unit_price"`
	DiscountPercent *float64  `db:"discount_percent"`
	TotalLine       float64   `db:"total_line"`
	CreatedAt       time.Time `db:"created_at"`
	UpdatedAt       time.Time `db:"updated_at"`
}

type CreateInvoiceItemRequest struct {
	InvoiceID       int
	ProductID       int
	Quantity        float64
	UnitPrice       float64
	DiscountPercent *float64
}

type UpdateInvoiceItemRequest struct {
	ID              int
	Quantity        float64
	UnitPrice       float64
	DiscountPercent *float64
}
