-- +goose Up
CREATE TABLE payments (
    month    TEXT    NOT NULL,
    child_id INTEGER NOT NULL REFERENCES children (id),
    PRIMARY KEY (month, child_id)
);

-- +goose Down
DROP TABLE payments;
