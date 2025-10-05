-- +goose Up
-- +goose StatementBegin
alter table account add column email varchar(255);
update account set email = concat(account.name, '@sluz.ch');
alter table account add constraint unique_email unique (email);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
alter table account drop column email;
-- +goose StatementEnd
