-- +goose Up
-- SQL in this section is executed when the migration is applied.
create schema if not exists storage;
create table if not exists storage.short_links
(
    id             serial       not null
        constraint short_links_pk
            primary key,
    user_id        varchar(50),
    origin         varchar(255) not null,
    short          varchar(50)  not null,
    correlation_id varchar(100)
);
comment on table storage.short_links is 'Short links from users';
comment on column storage.short_links.id is 'identifier of record';
comment on column storage.short_links.user_id is 'User identifier';
comment on column storage.short_links.origin is 'Origin link';
comment on column storage.short_links.short is 'Short link';
comment on column storage.short_links.correlation_id is 'Correlation identity';
alter table storage.short_links
    owner to postgres;
create unique index if not exists short_links_user_id_origin_uindex
    on storage.short_links (user_id, origin);


-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
drop table storage.short_links;
drop schema storage;



