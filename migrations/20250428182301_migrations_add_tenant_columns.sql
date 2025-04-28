-- +goose Up
-- +goose StatementBegin
alter table account add column tenant varchar(100);
alter table api_key add column tenant varchar(100);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
alter table account drop column tenant;
alter table api_key drop column tenant;
-- +goose StatementEnd
