-- +goose Up
-- +goose StatementBegin
create table if not exists account (
    id integer primary key generated always as identity,
    name varchar(100) not null,
    role varchar(50) not null,
    registered timestamptz not null default now(),
    password varchar(255)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table if exists account;
-- +goose StatementEnd
