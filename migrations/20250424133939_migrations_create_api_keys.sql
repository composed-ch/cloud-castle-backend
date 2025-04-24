-- +goose Up
-- +goose StatementBegin
create table if not exists api_key (
    id integer primary key generated always as identity,
    account_id integer not null references account(id) on delete cascade,
    zone varchar(100) null,
    api_key varchar(100) not null unique,
    api_secret varchar(100) not null
)
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table if exists api_key;
-- +goose StatementEnd
