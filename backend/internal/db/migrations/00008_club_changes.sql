-- +goose Up
CREATE TABLE club_changes (
    date     TEXT    NOT NULL,
    child_id INTEGER NOT NULL REFERENCES children (id),
    name     TEXT    NOT NULL,
    start    TEXT    NULL,
    PRIMARY KEY (date, child_id, name)
);

-- +goose Down
DROP TABLE club_changes;
