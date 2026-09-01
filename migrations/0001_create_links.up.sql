CREATE TABLE links (
    code         VARCHAR(8) PRIMARY KEY,
    original_url TEXT NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT links_code_format
        CHECK (code ~ '^[0-9A-Za-z]{8}$'),

    CONSTRAINT links_original_url_length
        CHECK (char_length(original_url) BETWEEN 1 AND 2048)
);
