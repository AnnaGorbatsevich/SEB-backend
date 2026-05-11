--liquibase formatted sql

--changeset 1:create_cursor_positions_table
CREATE TABLE IF NOT EXISTS cursor_positions (
    id          SERIAL PRIMARY KEY,
    session_id  TEXT             NOT NULL,
    x           DOUBLE PRECISION NOT NULL,
    y           DOUBLE PRECISION NOT NULL,
    ts          TIMESTAMPTZ      NOT NULL,
    received_at TIMESTAMPTZ      NOT NULL DEFAULT NOW(),
    email       TEXT             NOT NULL
);

--changeset 2:add_index_cursor_positions_session_id_email
CREATE INDEX IF NOT EXISTS idx_cursor_positions_session_email ON cursor_positions(session_id, email);

--changeset 3:create_key_presses_table
CREATE TABLE IF NOT EXISTS key_presses (
    id          SERIAL PRIMARY KEY,
    session_id  TEXT        NOT NULL,
    key_code    INTEGER     NOT NULL,
    key_name    TEXT        NOT NULL,
    modifiers   TEXT        NOT NULL DEFAULT '[]',
    is_combo    BOOLEAN     NOT NULL,
    ts          TIMESTAMPTZ NOT NULL,
    received_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    email       TEXT        NOT NULL
);

--changeset 4:add_index_key_presses_session_id_email
CREATE INDEX IF NOT EXISTS idx_key_presses_session_email ON key_presses(session_id, email);

--changeset 5:create_logs_table
CREATE TABLE IF NOT EXISTS logs (
    id          SERIAL PRIMARY KEY,
    session_id  TEXT        NOT NULL,
    level       TEXT        NOT NULL,
    message     TEXT        NOT NULL,
    ts          TIMESTAMPTZ NOT NULL,
    received_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    email       TEXT        NOT NULL
);

--changeset 6:add_index_logs_session_id_email
CREATE INDEX IF NOT EXISTS idx_logs_session_email ON logs(session_id, email);

--changeset 7:create_diagnostics_table
CREATE TABLE IF NOT EXISTS diagnostics (
    id          SERIAL PRIMARY KEY,
    session_id  TEXT        NOT NULL,
    code        TEXT        NOT NULL,
    status      TEXT        NOT NULL,
    details     JSONB       NOT NULL DEFAULT '{}',
    received_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    email       TEXT        NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_diagnostics_session_email ON diagnostics(session_id, email);
