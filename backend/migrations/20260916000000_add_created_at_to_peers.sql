-- +goose Up
-- +goose StatementBegin
ALTER TABLE peer_companies ADD COLUMN created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE peer_companies DROP COLUMN created_at;
-- +goose StatementEnd
