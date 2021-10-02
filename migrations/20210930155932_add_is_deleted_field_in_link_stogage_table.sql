-- +goose Up
-- SQL in this section is executed when the migration is applied.
alter table storage.short_links
    add is_deleted boolean default false;

comment on column storage.short_links.is_deleted is 'Is deleted link';


-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
alter table storage.short_links drop column is_deleted;

