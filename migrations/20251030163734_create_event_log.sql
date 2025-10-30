-- +goose Up
-- +goose StatementBegin
create table if not exists event_log (
    id integer primary key generated always as identity,
    account_id integer not null references account (id)
        on delete cascade,
    kind varchar(100) not null,
    happened timestamptz not null default now(),
    info_key varchar(100) null,
    info_val varchar(255) null
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table if exists event_log;
-- +goose StatementEnd
