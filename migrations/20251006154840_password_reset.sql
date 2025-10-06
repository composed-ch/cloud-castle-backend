-- +goose Up
-- +goose StatementBegin
create table if not exists password_reset (
    id integer primary key generated always as identity,
    account_id integer not null references account (id)
        on delete cascade,
    token varchar(255) not null,
    created timestamptz not null default now(),
    expires timestamptz not null default now() + interval '30 minutes'
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table if exists password_reset;
-- +goose StatementEnd
