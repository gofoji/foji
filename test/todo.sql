-- migrate:up
DROP TABLE IF EXISTS todo;
DROP TABLE IF EXISTS category;

CREATE TABLE category
    (
        id SERIAL NOT NULL
            CONSTRAINT category_pkey
                PRIMARY KEY,
        label VARCHAR
    );

CREATE TABLE todo
    (
        id SERIAL NOT NULL
            CONSTRAINT todo_pkey
                PRIMARY KEY,
        position INT4,
        label VARCHAR NOT NULL,
        label_nullable VARCHAR,
        minutes INTEGER NOT NULL,
        minutes_nullable INTEGER,
        parameters JSONB,
        category_id INTEGER
            CONSTRAINT todo_category_id_fkey REFERENCES category,
        expires_at TIMESTAMP WITH TIME ZONE,
        due_at TIMESTAMP WITH TIME ZONE,
        completed_at TIMESTAMP WITH TIME ZONE,
        updated_at TIMESTAMP WITH TIME ZONE,
        deleted_at TIMESTAMP WITH TIME ZONE
    );

CREATE UNIQUE INDEX idx_todo_label ON todo (label);
CREATE INDEX idx_todo_category ON todo (category_id);

INSERT INTO
    public.category (label)
VALUES
    ('category 1'),
    ('category 2');

INSERT INTO
    public.todo (label, minutes, category_id, expires_at)
VALUES
    ('Foo', 1, 1, NULL),
    ('Bar', 2, 2, '2020-07-03 11:54:46.786000'),
    ('Funky', 3, NULL, NULL);

-- migrate:down

DROP TABLE IF EXISTS todo;
DROP TABLE IF EXISTS category;
