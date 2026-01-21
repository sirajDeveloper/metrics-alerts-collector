CREATE TABLE metrics (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    type TEXT NOT NULL,
    delta BIGINT,
    value DOUBLE PRECISION
);

CREATE INDEX metrics_name_idx ON metrics (name);
CREATE INDEX metrics_type_idx ON metrics (type);