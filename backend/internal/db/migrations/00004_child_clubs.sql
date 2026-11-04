-- +goose Up
CREATE TABLE child_clubs (
    child_id INTEGER NOT NULL REFERENCES children (id),
    weekday  INTEGER NOT NULL CHECK (weekday BETWEEN 1 AND 7),
    start    TEXT    NOT NULL,
    name     TEXT    NOT NULL,
    PRIMARY KEY (child_id, weekday, start, name)
);

-- +goose Down
DROP TABLE child_clubs;
