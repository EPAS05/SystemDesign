CREATE TABLE IF NOT EXISTS units (
    id          SERIAL PRIMARY KEY,
    name        TEXT NOT NULL,
    multiplier  DOUBLE PRECISION NOT NULL,
    created_at  TIMESTAMPTZ DEFAULT now(),
    updated_at  TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE IF NOT EXISTS classifier_nodes (
    id              SERIAL PRIMARY KEY,
    name            TEXT NOT NULL,
    parent_id       INTEGER REFERENCES classifier_nodes(id) ON DELETE RESTRICT,
    is_terminal     BOOLEAN,                     
    unit_id         INTEGER REFERENCES units(id) ON DELETE SET NULL,
    sort_order      INTEGER DEFAULT 0,
    created_at      TIMESTAMPTZ DEFAULT now(),
    updated_at      TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE IF NOT EXISTS products (
    id              SERIAL PRIMARY KEY,
    name            TEXT NOT NULL,
    class_node_id   INTEGER NOT NULL REFERENCES classifier_nodes(id) ON DELETE RESTRICT,
    unit_type       TEXT CHECK (unit_type IN ('mass', 'length', 'piece')),
    weight_per_meter DOUBLE PRECISION,
    piece_length    DOUBLE PRECISION,
    default_unit_id INTEGER REFERENCES units(id) ON DELETE SET NULL,
    created_at      TIMESTAMPTZ DEFAULT now(),
    updated_at      TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE IF NOT EXISTS enums (
    id          SERIAL PRIMARY KEY,
    name        TEXT NOT NULL UNIQUE,
    description TEXT,
    type_node_id   INTEGER NOT NULL REFERENCES classifier_nodes(id) ON DELETE RESTRICT,
    created_at  TIMESTAMPTZ DEFAULT now(),
    updated_at  TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE IF NOT EXISTS enum_values (
    id          SERIAL PRIMARY KEY,
    enum_id     INTEGER REFERENCES enums(id) ON DELETE CASCADE,
    value       TEXT NOT NULL,
    sort_order  INTEGER DEFAULT 0,
    created_at  TIMESTAMPTZ DEFAULT now(),
    updated_at  TIMESTAMPTZ DEFAULT now(),
    UNIQUE(enum_id, value)
);

CREATE TABLE IF NOT EXISTS parameter_definitions (
    id              SERIAL PRIMARY KEY,
    class_node_id   INTEGER NOT NULL REFERENCES classifier_nodes(id) ON DELETE CASCADE,
    name            TEXT NOT NULL,
    description     TEXT,
    parameter_type  TEXT NOT NULL CHECK (parameter_type IN ('number', 'enum')),
    unit_id         INTEGER REFERENCES units(id) ON DELETE SET NULL,
    enum_id         INTEGER REFERENCES enums(id) ON DELETE SET NULL,
    is_required     BOOLEAN DEFAULT false,
    sort_order      INTEGER DEFAULT 0,
    created_at      TIMESTAMPTZ DEFAULT now(),
    updated_at      TIMESTAMPTZ DEFAULT now(),
    UNIQUE(class_node_id, name)
);

CREATE TABLE IF NOT EXISTS parameter_constraints (
    id              SERIAL PRIMARY KEY,
    param_def_id    INTEGER NOT NULL REFERENCES parameter_definitions(id) ON DELETE CASCADE,
    min_value       DOUBLE PRECISION,
    max_value       DOUBLE PRECISION,
    created_at      TIMESTAMPTZ DEFAULT now(),
    updated_at      TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE IF NOT EXISTS parameter_values (
    id              SERIAL PRIMARY KEY,
    product_id      INTEGER NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    param_def_id    INTEGER NOT NULL REFERENCES parameter_definitions(id) ON DELETE CASCADE,
    value_numeric   DOUBLE PRECISION,
    value_enum_id   INTEGER REFERENCES enum_values(id) ON DELETE SET NULL,
    created_at      TIMESTAMPTZ DEFAULT now(),
    updated_at      TIMESTAMPTZ DEFAULT now(),
    UNIQUE(product_id, param_def_id)
);

CREATE TABLE IF NOT EXISTS customers (
    id          SERIAL PRIMARY KEY,
    name        TEXT NOT NULL,
    tax_id      TEXT UNIQUE,      -- ИНН (уникальный, если заполнен)
    address     TEXT,
    created_at  TIMESTAMPTZ DEFAULT now(),
    updated_at  TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE IF NOT EXISTS invoices (
    id              SERIAL PRIMARY KEY,
    invoice_number  TEXT NOT NULL UNIQUE,
    invoice_date    DATE NOT NULL DEFAULT CURRENT_DATE,
    invoice_type    TEXT NOT NULL CHECK (invoice_type IN ('incoming', 'outgoing', 'return')),
    status          TEXT NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'confirmed', 'paid', 'shipped', 'cancelled')),
    customer_id     INTEGER NOT NULL REFERENCES customers(id) ON DELETE RESTRICT,
    currency        TEXT NOT NULL DEFAULT 'RUB' CHECK (currency IN ('RUB', 'USD', 'EUR')),
    total_amount    NUMERIC(15,2) NOT NULL DEFAULT 0,
    discount_total  NUMERIC(15,2) DEFAULT 0,
    tax_rate        NUMERIC(5,2) DEFAULT 0,
    tax_amount      NUMERIC(15,2) DEFAULT 0,
    comment         TEXT,
    created_at      TIMESTAMPTZ DEFAULT now(),
    updated_at      TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE IF NOT EXISTS invoice_items (
    id               SERIAL PRIMARY KEY,
    invoice_id       INTEGER NOT NULL REFERENCES invoices(id) ON DELETE CASCADE,
    product_id       INTEGER NOT NULL REFERENCES products(id) ON DELETE RESTRICT,
    quantity         NUMERIC(15,3) NOT NULL CHECK (quantity > 0),
    unit_price       NUMERIC(15,2) NOT NULL,
    discount_percent NUMERIC(5,2) DEFAULT 0,
    total_line       NUMERIC(15,2) NOT NULL,
    created_at       TIMESTAMPTZ DEFAULT now(),
    updated_at       TIMESTAMPTZ DEFAULT now()
);

INSERT INTO classifier_nodes (id, name, parent_id, is_terminal, unit_id, sort_order)
VALUES (1, 'Trash', NULL, false, NULL, 0),
       (2, 'Изделия', NULL, false, NULL, 0),
       (3, 'Перечисления', NULL, false, NULL, 0)
ON CONFLICT (id) DO NOTHING;

INSERT INTO classifier_nodes (id, name, parent_id, is_terminal, unit_id, sort_order)
VALUES (4, 'Числовые', 3, false, NULL, 0),
       (5, 'Строковые', 3, false, NULL, 0),
       (6, 'Картинки', 3, false, NULL, 0)
ON CONFLICT (id) DO NOTHING;

INSERT INTO units (name, multiplier) VALUES
    ('метр', 1.0),
    ('миллиметр', 0.001),
    ('сантиметр', 0.01),
    ('километр', 1000.0),
    ('тонна', 1.0),
    ('килограмм', 0.001),
    ('грамм', 0.000001),
    ('штука', 1.0)
ON CONFLICT DO NOTHING;

SELECT setval('classifier_nodes_id_seq', COALESCE((SELECT MAX(id) FROM classifier_nodes), 0));
SELECT setval('units_id_seq', COALESCE((SELECT MAX(id) FROM units), 0));
SELECT setval('products_id_seq', COALESCE((SELECT MAX(id) FROM products), 0));
SELECT setval('enums_id_seq', COALESCE((SELECT MAX(id) FROM enums), 0));
