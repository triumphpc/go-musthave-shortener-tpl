package db

import (
	"context"
	"errors"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/helpers"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/helpers/db"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/models/shortlink"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/models/user"
)

// PostgreSQLStorage storage
type PostgreSQLStorage struct{}

// ErrURLNotFound error by package level
var ErrURLNotFound = errors.New("url not found")

// New Instance new Storage with not null fields
func New() (*PostgreSQLStorage, error) {
	// Check if scheme exist
	sql := `
create schema if not exists storage;
create table if not exists storage.short_links(
id      serial       not null
constraint short_links_pk
primary key,
	user_id varchar(50),
	origin  varchar(255) not null,
	short   varchar(50)  not null
);
comment on table storage.short_links is 'Short links from users';
comment on column storage.short_links.id is 'identifier of record';
comment on column storage.short_links.user_id is 'User identifier';
comment on column storage.short_links.origin is 'Origin link';
comment on column storage.short_links.short is 'Short link';
alter table storage.short_links
owner to postgres;
create unique index if not exists short_links_user_id_origin_uindex
on storage.short_links (user_id, origin);
`
	if err := db.Insert(context.Background(), sql); err != nil {
		return &PostgreSQLStorage{}, err
	}
	return &PostgreSQLStorage{}, nil
}

// LinkByShort implement interface for get data from storage by userId and shortLink
func (s *PostgreSQLStorage) LinkByShort(short shortlink.Short) (string, error) {
	sql := "select origin from storage.short_links where short=$1"
	dbd, _ := db.Instance()

	var origin string
	err := dbd.QueryRowContext(context.Background(), sql, string(short)).Scan(&origin)
	if err != nil {
		return "", ErrURLNotFound
	}
	return origin, nil
}

// LinksByUser return all user links
func (s *PostgreSQLStorage) LinksByUser(userID user.UniqUser) (shortlink.ShortLinks, error) {
	sql := "select origin, short from storage.short_links where user_id=$1"
	dbd, _ := db.Instance()

	origins := shortlink.ShortLinks{}
	rows, err := dbd.QueryContext(context.Background(), sql, string(userID))
	if err != nil {
		return origins, err
	}

	for rows.Next() {
		var origin string
		var short string
		err = rows.Scan(&origin, &short)
		if err != nil {
			return origins, err
		}
		origins[shortlink.Short(short)] = origin
	}
	return origins, nil
}

// Save url in storage of short links
func (s *PostgreSQLStorage) Save(userID user.UniqUser, url string) shortlink.Short {
	short := shortlink.Short(helpers.RandomString(10))
	// Save to database
	sql := `
insert into storage.short_links (id, user_id, origin, short) 
values (default, $1, $2, $3)
on conflict (user_id, origin)
do nothing;
`
	err := db.Insert(context.Background(), sql, userID, url, short)
	if err != nil {
		panic(err)
	}

	return short
}
