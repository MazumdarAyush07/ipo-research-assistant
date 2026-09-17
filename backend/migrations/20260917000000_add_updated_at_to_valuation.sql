-- +goose Up
-- +goose StatementBegin
ALTER TABLE valuation ADD COLUMN updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE valuation DROP COLUMN IF EXISTS updated_at;
-- +goose StatementEnd
