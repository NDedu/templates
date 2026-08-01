CREATE TABLE IF NOT EXISTS books (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    image TEXT DEFAULT '',
    uuid TEXT NOT NULL DEFAULT (lower(hex(randomblob(4))) || '-' || lower(hex(randomblob(2))) || '-4' || substr(lower(hex(randomblob(2))),2) || '-' || substr('89ab',abs(random()) % 4 + 1, 1) || substr(lower(hex(randomblob(2))),2) || '-' || lower(hex(randomblob(6)))),
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX idx_books_uuid ON books (uuid);

INSERT INTO books (name, description, image)
VALUES 
    ('Lord of the Rings', 'Greatest book of all times!', '/static/img/lotr.png'),
    ('The Hobbit', 'Prequal to the greatest book of all times.', '/static/img/hobbit.png'),
    ('Silmarillion', 'Lore of the greatest book of all times.', '/static/img/silmarillion.png');