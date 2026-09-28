-- +goose Up
ALTER TABLE attendance ADD COLUMN status_new TEXT NOT NULL DEFAULT 'pending'
    CHECK (status_new IN ('pending', 'present', 'absent'));
UPDATE attendance SET status_new = status;
ALTER TABLE attendance DROP COLUMN status;
ALTER TABLE attendance RENAME COLUMN status_new TO status;

CREATE TABLE day_clubs (
    date     TEXT    NOT NULL,
    child_id INTEGER NOT NULL REFERENCES children (id),
    start    TEXT    NOT NULL,
    name     TEXT    NOT NULL,
    PRIMARY KEY (date, child_id, start, name)
);

-- +goose Down
DROP TABLE day_clubs;

UPDATE attendance SET status = 'present', price = (SELECT price FROM children WHERE id = child_id)
WHERE status = 'pending';
ALTER TABLE attendance ADD COLUMN status_old TEXT NOT NULL DEFAULT 'present'
    CHECK (status_old IN ('present', 'absent'));
UPDATE attendance SET status_old = status;
ALTER TABLE attendance DROP COLUMN status;
ALTER TABLE attendance RENAME COLUMN status_old TO status;
