-- +goose Up
-- +goose StatementBegin
alter table api_key drop column account_id;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
alter table api_key add column account_id integer null references account(id) on delete cascade;
-- +goose StatementEnd
