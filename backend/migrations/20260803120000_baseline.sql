-- +goose Up
-- +goose StatementBegin
SELECT 'baseline';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'baseline down';
-- +goose StatementEnd
