package db

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"github.com/jackc/pgerrcode"
	"github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/helpers"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/models/shortlink"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/models/user"
	"go.uber.org/zap"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

// PostgreSQLStorage storage
type PostgreSQLStorage struct {
	db *sql.DB
	l  *zap.Logger
}

// ErrURLNotFound error by package level
var ErrURLNotFound = errors.New("url not found")

// ErrAlreadyHasShort if exist
var ErrAlreadyHasShort = errors.New("already has short")

// Scheme of database
//const scheme = `
//create schema if not exists storage;
//create table if not exists storage.short_links
//(
//    id             serial       not null
//        constraint short_links_pk
//            primary key,
//    user_id        varchar(50),
//    origin         varchar(255) not null,
//    short          varchar(50)  not null,
//    correlation_id varchar(100)
//);
//comment on table storage.short_links is 'Short links from users';
//comment on column storage.short_links.id is 'identifier of record';
//comment on column storage.short_links.user_id is 'User identifier';
//comment on column storage.short_links.origin is 'Origin link';
//comment on column storage.short_links.short is 'Short link';
//comment on column storage.short_links.correlation_id is 'Correlation identity';
//alter table storage.short_links
//    owner to postgres;
//create unique index if not exists short_links_user_id_origin_uindex
//    on storage.short_links (user_id, origin);
//`

// sqlNewRecord for new record in db
const sqlNewRecord = `
insert into storage.short_links (id, user_id, origin, short) 
values (default, $1, $2, $3)
`
const sqlGetCurrentRecord = `
select short from storage.short_links where user_id=$1 and origin=$2
`

// sqlBunchNewRecord for new record in db
const sqlBunchNewRecord = `
insert into storage.short_links (id, user_id, origin, short, correlation_id) 
values (default, $1, $2, $3, $4)
on conflict (user_id, origin)
do nothing;
`

// sqlSelectFromOrigin select origin
const sqlSelectOrigin = `
select origin from storage.short_links where short=$1
`

// sqlSelectOriginAndShort select origin and short
const sqlSelectOriginAndShort = `
select origin, short from storage.short_links where user_id=$1
`

// New New new Storage with not null fields
func New(c *sql.DB, l *zap.Logger) (*PostgreSQLStorage, error) {
	// Check if scheme exist
	goose.SetBaseFS(embedMigrations)
	if err := goose.Up(c, "migrations"); err != nil {
		panic(err)
	}

	//if _, err := c.ExecContext(context.Background(), scheme); err != nil {
	//	return nil, err
	//}
	return &PostgreSQLStorage{c, l}, nil
}

// LinkByShort implement interface for get data from storage by userId and shortLink
func (s *PostgreSQLStorage) LinkByShort(short shortlink.Short) (string, error) {
	var origin string
	err := s.db.QueryRowContext(context.Background(), sqlSelectOrigin, string(short)).Scan(&origin)

	if err != nil {
		return "", ErrURLNotFound
	}
	return origin, nil
}

// LinksByUser return all user links
func (s *PostgreSQLStorage) LinksByUser(userID user.UniqUser) (shortlink.ShortLinks, error) {
	origins := shortlink.ShortLinks{}
	rows, err := s.db.QueryContext(context.Background(), sqlSelectOriginAndShort, string(userID))
	if err != nil {
		return origins, err
	}

	err = rows.Err()
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
func (s *PostgreSQLStorage) Save(userID user.UniqUser, origin string) (shortlink.Short, error) {
	short := shortlink.Short(helpers.RandomString(10))
	// Save to database
	if _, err := s.db.ExecContext(context.Background(), sqlNewRecord, userID, origin, short); err != nil {
		if err, ok := err.(*pq.Error); ok {
			if err.Code == pgerrcode.UniqueViolation {
				// take current link
				var short string
				_ = s.db.QueryRowContext(context.Background(), sqlGetCurrentRecord, string(userID), origin).Scan(&short)
				return shortlink.Short(short), ErrAlreadyHasShort
			}
		}
		return short, err
	}
	return short, nil
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
	var shorts []shortlink.ShortURLs
	// Delete old records for tests
	_, _ = s.db.Exec("truncate table storage.short_links;")
	// Start transaction
	tx, err := s.db.Begin()
	if err != nil {
		return shorts, err
	}
	// Rollback handler
	defer func(tx *sql.Tx) {
		_ = tx.Rollback()
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
			s.l.Info("Close statement error", zap.Error(err))
		}
	}(stmt)

	for _, v := range buffer {
		// Add record to transaction
		if _, err = stmt.ExecContext(context.Background(), "all", v.Origin, v.Short, v.ID); err == nil {
			shorts = append(shorts, shortlink.ShortURLs{
				Short: v.Short,
				ID:    v.ID,
			})
		} else {
			s.l.Info("Save bunch error", zap.Error(err))
		}
	}
	// шаг 4 — сохраняем изменения
	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return shorts, nil
}
