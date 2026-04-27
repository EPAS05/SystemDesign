BEGIN;

-- Test classifier tree under node 2 ("Изделия")
INSERT INTO classifier_nodes (id, name, parent_id, is_terminal, unit_id, sort_order)
VALUES
    (1000, 'Металлопрокат', 2, false, NULL, 10),
    (1001, 'Балки', 1000, false, NULL, 20),
    (1002, 'Двутавровые балки', 1001, true, NULL, 30),
    (1003, 'Швеллеры', 1001, true, NULL, 40),
    (1004, 'Листовой прокат', 1000, false, NULL, 50),
    (1005, 'Листы горячекатаные', 1004, true, NULL, 60)
ON CONFLICT (id) DO UPDATE
SET
    name = EXCLUDED.name,
    parent_id = EXCLUDED.parent_id,
    is_terminal = EXCLUDED.is_terminal,
    unit_id = EXCLUDED.unit_id,
    sort_order = EXCLUDED.sort_order,
    updated_at = now();

-- Enum in "Строковые" branch (node 5)
INSERT INTO enums (id, name, description, type_node_id)
VALUES
    (1000, 'Марка стали (seed)', 'Тестовые марки стали', 5)
ON CONFLICT (id) DO UPDATE
SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    type_node_id = EXCLUDED.type_node_id,
    updated_at = now();

INSERT INTO enum_values (id, enum_id, value, sort_order)
VALUES
    (1000, 1000, 'Ст3', 0),
    (1001, 1000, '09Г2С', 1),
    (1002, 1000, 'S355', 2)
ON CONFLICT (id) DO UPDATE
SET
    enum_id = EXCLUDED.enum_id,
    value = EXCLUDED.value,
    sort_order = EXCLUDED.sort_order,
    updated_at = now();

-- Parameter definitions for terminal classes
INSERT INTO parameter_definitions (
    id,
    class_node_id,
    name,
    description,
    parameter_type,
    unit_id,
    enum_id,
    is_required,
    sort_order
)
VALUES
    (1000, 1001, 'Длина', 'Длина профиля', 'number', 1, NULL, true, 10),
    (1001, 1002, 'Марка стали', 'Марка стали балки', 'enum', NULL, 1000, true, 20),
    (1002, 1003, 'Высота профиля', 'Высота швеллера', 'number', 2, NULL, true, 10),
    (1003, 1003, 'Марка стали', 'Марка стали швеллера', 'enum', NULL, 1000, true, 20)
ON CONFLICT (id) DO UPDATE
SET
    class_node_id = EXCLUDED.class_node_id,
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    parameter_type = EXCLUDED.parameter_type,
    unit_id = EXCLUDED.unit_id,
    enum_id = EXCLUDED.enum_id,
    is_required = EXCLUDED.is_required,
    sort_order = EXCLUDED.sort_order,
    updated_at = now();

INSERT INTO parameter_constraints (id, param_def_id, min_value, max_value)
VALUES
    (1000, 1000, 1, 24),
    (1001, 1002, 80, 500)
ON CONFLICT (id) DO UPDATE
SET
    param_def_id = EXCLUDED.param_def_id,
    min_value = EXCLUDED.min_value,
    max_value = EXCLUDED.max_value,
    updated_at = now();

-- Products in terminal classes
INSERT INTO products (
    id,
    name,
    class_node_id,
    unit_type,
    weight_per_meter,
    piece_length,
    default_unit_id
)
VALUES
    (1000, 'Балка 20Б1', 1002, 'mass', 0.021, 12.0, 5),
    (1001, 'Балка 25Б1', 1002, 'mass', 0.0257, 12.0, 5),
    (1002, 'Швеллер 16П', 1003, 'mass', 0.0142, 12.0, 5),
    (1003, 'Лист 10x1500x6000', 1005, 'piece', NULL, 6.0, 8)
ON CONFLICT (id) DO UPDATE
SET
    name = EXCLUDED.name,
    class_node_id = EXCLUDED.class_node_id,
    unit_type = EXCLUDED.unit_type,
    weight_per_meter = EXCLUDED.weight_per_meter,
    piece_length = EXCLUDED.piece_length,
    default_unit_id = EXCLUDED.default_unit_id,
    updated_at = now();

-- Parameter values for products
INSERT INTO parameter_values (id, product_id, param_def_id, value_numeric, value_enum_id)
VALUES
    (1000, 1000, 1000, 12.0, NULL),
    (1001, 1000, 1001, NULL, 1001),
    (1002, 1001, 1000, 12.0, NULL),
    (1003, 1001, 1001, NULL, 1002),
    (1004, 1002, 1002, 160.0, NULL),
    (1005, 1002, 1003, NULL, 1000)
ON CONFLICT (id) DO UPDATE
SET
    product_id = EXCLUDED.product_id,
    param_def_id = EXCLUDED.param_def_id,
    value_numeric = EXCLUDED.value_numeric,
    value_enum_id = EXCLUDED.value_enum_id,
    updated_at = now();

-- Customer and one invoice
INSERT INTO customers (id, name, tax_id, address)
VALUES
    (1000, 'ООО Тест-Клиент', '7701000000', 'Москва, ул. Тестовая, 1')
ON CONFLICT (id) DO UPDATE
SET
    name = EXCLUDED.name,
    tax_id = EXCLUDED.tax_id,
    address = EXCLUDED.address,
    updated_at = now();

INSERT INTO invoices (
    id,
    invoice_number,
    invoice_date,
    invoice_type,
    status,
    customer_id,
    currency,
    total_amount,
    discount_total,
    tax_rate,
    tax_amount,
    comment
)
VALUES
    (
        1000,
        'SEED-2026-0001',
        CURRENT_DATE,
        'outgoing',
        'confirmed',
        1000,
        'RUB',
        204000.00,
        0,
        20.00,
        34000.00,
        'Тестовый счет из seed.sql'
    )
ON CONFLICT (id) DO UPDATE
SET
    invoice_number = EXCLUDED.invoice_number,
    invoice_date = EXCLUDED.invoice_date,
    invoice_type = EXCLUDED.invoice_type,
    status = EXCLUDED.status,
    customer_id = EXCLUDED.customer_id,
    currency = EXCLUDED.currency,
    total_amount = EXCLUDED.total_amount,
    discount_total = EXCLUDED.discount_total,
    tax_rate = EXCLUDED.tax_rate,
    tax_amount = EXCLUDED.tax_amount,
    comment = EXCLUDED.comment,
    updated_at = now();

INSERT INTO invoice_items (
    id,
    invoice_id,
    product_id,
    quantity,
    unit_price,
    discount_percent,
    total_line
)
VALUES
    (1000, 1000, 1000, 10.000, 12000.00, 0, 120000.00),
    (1001, 1000, 1002, 8.000, 10500.00, 0, 84000.00)
ON CONFLICT (id) DO UPDATE
SET
    invoice_id = EXCLUDED.invoice_id,
    product_id = EXCLUDED.product_id,
    quantity = EXCLUDED.quantity,
    unit_price = EXCLUDED.unit_price,
    discount_percent = EXCLUDED.discount_percent,
    total_line = EXCLUDED.total_line,
    updated_at = now();

-- Keep sequences in sync after manual IDs
SELECT setval('classifier_nodes_id_seq', COALESCE((SELECT MAX(id) FROM classifier_nodes), 1));
SELECT setval('products_id_seq', COALESCE((SELECT MAX(id) FROM products), 1));
SELECT setval('enums_id_seq', COALESCE((SELECT MAX(id) FROM enums), 1));
SELECT setval('enum_values_id_seq', COALESCE((SELECT MAX(id) FROM enum_values), 1));
SELECT setval('parameter_definitions_id_seq', COALESCE((SELECT MAX(id) FROM parameter_definitions), 1));
SELECT setval('parameter_constraints_id_seq', COALESCE((SELECT MAX(id) FROM parameter_constraints), 1));
SELECT setval('parameter_values_id_seq', COALESCE((SELECT MAX(id) FROM parameter_values), 1));
SELECT setval('customers_id_seq', COALESCE((SELECT MAX(id) FROM customers), 1));
SELECT setval('invoices_id_seq', COALESCE((SELECT MAX(id) FROM invoices), 1));
SELECT setval('invoice_items_id_seq', COALESCE((SELECT MAX(id) FROM invoice_items), 1));

COMMIT;
