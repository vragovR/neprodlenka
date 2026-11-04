-- +goose Up
ALTER TABLE report_months ADD COLUMN message_id TEXT;

-- +goose Down
ALTER TABLE report_months DROP COLUMN message_id;
