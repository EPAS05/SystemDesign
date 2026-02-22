CREATE TABLE IF NOT EXISTS units (
    id          SERIAL PRIMARY KEY,
    name        TEXT NOT NULL,
    multiplier  DOUBLE PRECISION NOT NULL,
    created_at  TIMESTAMPTZ DEFAULT now(),
    updated_at  TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE IF NOT EXISTS classifier_nodes (
    id          SERIAL PRIMARY KEY,
    name        TEXT NOT NULL,
    parent_id   INTEGER REFERENCES classifier_nodes(id) ON DELETE RESTRICT,
    node_type   TEXT NOT NULL CHECK (node_type IN ('metaclass', 'leaf')),
    is_terminal BOOLEAN,
    unit_id     INTEGER REFERENCES units(id) ON DELETE SET NULL,
    sort_order  INTEGER DEFAULT 0,
    created_at  TIMESTAMPTZ DEFAULT now(),
    updated_at  TIMESTAMPTZ DEFAULT now()
);


INSERT INTO classifier_nodes (id, name, parent_id, node_type, is_terminal, unit_id, sort_order)
VALUES (1, 'Trash', NULL, 'metaclass', false, NULL, 0)
ON CONFLICT (id) DO NOTHING;

INSERT INTO units (name, multiplier) VALUES
    ('метр', 1.0),
    ('миллиметр', 0.001),
    ('сантиметр', 0.01),
    ('километр', 1000.0)
ON CONFLICT DO NOTHING;