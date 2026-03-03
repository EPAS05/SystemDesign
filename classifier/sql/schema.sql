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
    node_type       TEXT NOT NULL CHECK (node_type IN ('metaclass', 'leaf')),
    is_terminal     BOOLEAN,
    unit_id         INTEGER REFERENCES units(id) ON DELETE SET NULL,
    sort_order      INTEGER DEFAULT 0,
    unit_type       TEXT CHECK (unit_type IN ('mass', 'length', 'piece')),
    weight_per_meter DOUBLE PRECISION,
    piece_length    DOUBLE PRECISION,
    default_unit_id INTEGER REFERENCES units(id) ON DELETE SET NULL,
    created_at      TIMESTAMPTZ DEFAULT now(),
    updated_at      TIMESTAMPTZ DEFAULT now()
);

INSERT INTO classifier_nodes (id, name, parent_id, node_type, is_terminal, unit_id, sort_order,
                              unit_type, weight_per_meter, piece_length, default_unit_id)
VALUES (1, 'Trash', NULL, 'metaclass', false, NULL, 0,
        NULL, NULL, NULL, NULL)
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

CREATE TABLE IF NOT EXISTS enums (
    id          SERIAL PRIMARY KEY,
    name        TEXT NOT NULL UNIQUE,            
    description TEXT,                            
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
)