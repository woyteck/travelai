CREATE TABLE conversations (
    id SERIAL PRIMARY KEY,
    uuid UUID NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP
);
CREATE INDEX idx_uuid ON conversations (uuid);

CREATE TABLE messages (
    id SERIAL PRIMARY KEY,
    conversation_id INT REFERENCES conversations(id),
    role VARCHAR(50),
    content TEXT
);

CREATE TYPE memory_type AS ENUM ('web_article', 'text_file');

CREATE TABLE memories (
    id SERIAL PRIMARY KEY,
    uuid UUID NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP,
    memory_type memory_type NOT NULL,
    source varchar(1024),
    content TEXT
);
CREATE INDEX idx_memory_uuid ON memories (uuid);

CREATE TABLE memory_fragments (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    content_original TEXT,
    content_refined TEXT,
    is_refined BOOLEAN DEFAULT false,
    is_embedded BOOLEAN DEFAULT false,
    memory_id INT REFERENCES memories(id) ON DELETE CASCADE
);

CREATE TABLE cache (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    valid_until TIMESTAMP,
    cache_key VARCHAR(1024),
    cache_value TEXT
);
CREATE INDEX idx_cache_key ON cache (cache_key);
CREATE INDEX idx_cache_valid_until ON cache (valid_until);
