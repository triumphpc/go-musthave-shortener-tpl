package db

import (
	"context"
	"database/sql"
	"errors"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/helpers"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/helpers/db"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/logger"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/models/shortlink"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/models/user"
	"go.uber.org/zap"
)

// PostgreSQLStorage storage
type PostgreSQLStorage struct{}

// ErrURLNotFound error by package level
var ErrURLNotFound = errors.New("url not found")

// Scheme of database
const scheme = `
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
comment on column storage.short_links.correlation_id is 'Correlation itentity';
alter table storage.short_links
    owner to postgres;
create unique index if not exists short_links_user_id_origin_uindex
    on storage.short_links (user_id, origin);
`

// sqlNewRecord for new record in db
const sqlNewRecord = `
insert into storage.short_links (id, user_id, origin, short) 
values (default, $1, $2, $3)
on conflict (user_id, origin)
do nothing;
`

// sqlBunchNewRecord for new record in db
const sqlBunchNewRecord = `
insert into storage.short_links (id, user_id, origin, short, correlation_id) 
values (default, $1, $2, $3, $4)
on conflict (user_id, origin)
do nothing;
`

// New Instance new Storage with not null fields
func New() (*PostgreSQLStorage, error) {
	// Check if scheme exist
	if err := db.Insert(context.Background(), scheme); err != nil {
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
	err := db.Insert(context.Background(), sqlNewRecord, userID, url, short)
	if err != nil {
		panic(err)
	}
	return short
}

// BunchSave save mass urls
func (s *PostgreSQLStorage) BunchSave(urls []shortlink.URLs) ([]shortlink.ShortURLs, error) {
	// Generate shorts
	type temp struct {
		ID,
		Origin,
		Short string
	}
	var buffer []temp
	for _, v := range urls {
		var t = temp{
			ID:     v.ID,
			Origin: v.Origin,
			Short:  helpers.RandomString(10),
		}
		buffer = append(buffer, t)
	}
	dbd, _ := db.Instance()
	var shorts []shortlink.ShortURLs
	// Start transaction
	tx, err := dbd.Begin()
	if err != nil {
		return shorts, err
	}
	// Rollback handler
	defer func(tx *sql.Tx) {
		err := tx.Rollback()
		if err != nil {
			logger.Info("Rollback error", zap.Error(err))
		}
	}(tx)
	// Prepare statement
	stmt, err := tx.PrepareContext(context.Background(), sqlBunchNewRecord)
	if err != nil {
		return shorts, err
	}
	// Close statement
	defer func(stmt *sql.Stmt) {
		err := stmt.Close()
		if err != nil {
			logger.Info("Close statement error", zap.Error(err))
		}
	}(stmt)

	for _, v := range buffer {
		// Add record to transaction
		if _, err = stmt.ExecContext(context.Background(), "all", v.Origin, v.Short, v.ID); err != nil {
			if err != nil {
				logger.Info("Save bunch error", zap.Error(err))
			} else {
				shorts = append(shorts, shortlink.ShortURLs{
					Short: v.Short,
					ID:    v.ID,
				})
			}
		}
	}
	// шаг 4 — сохраняем изменения
	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return shorts, nil
}
