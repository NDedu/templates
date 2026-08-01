CREATE TABLE IF NOT EXISTS books (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    image VARCHAR(255) DEFAULT '',
    uuid UUID NOT NULL DEFAULT gen_random_uuid(),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_books_uuid ON books (uuid);

INSERT INTO books (name, description)
VALUES ('Lord of the Rings', 'Greatest book of all times!');
