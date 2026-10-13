-- +goose Up
ALTER TABLE children ADD COLUMN weekdays INTEGER NOT NULL DEFAULT 31 CHECK (weekdays BETWEEN 1 AND 127);

-- +goose Down
ALTER TABLE children DROP COLUMN weekdays;
