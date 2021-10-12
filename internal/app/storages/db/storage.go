package db

import (
	"context"
	"database/sql"
	"github.com/jackc/pgerrcode"
	"github.com/lib/pq"
	"github.com/pressly/goose/v3"
	er "github.com/triumphpc/go-musthave-shortener-tpl/internal/app/errors"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/helpers"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/models/shortlink"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/models/user"
	"github.com/triumphpc/go-musthave-shortener-tpl/migrations"
	"go.uber.org/zap"
)

// PostgreSQLStorage storage
type PostgreSQLStorage struct {
	db *sql.DB
	l  *zap.Logger
}

// sqlNewRecord for new record in db
const sqlNewRecord = `
insert into storage.short_links (id, user_id, origin, short) 
values (default, $1, $2, $3)
`

// sqlDeleteRecords sql for clear records
const sqlDeleteRecords = `truncate storage.short_links`

// sqlGetCurrentRecord for get current record
const sqlGetCurrentRecord = "select short from storage.short_links where user_id=$1 and origin=$2;"

// sqlBunchNewRecord for new record in db
const sqlBunchNewRecord = `
insert into storage.short_links (id, user_id, origin, short, correlation_id) 
values (default, $1, $2, $3, $4)
on conflict (user_id, origin)
do nothing;
`

// sqlSelectFromOrigin select origin
const sqlSelectOrigin = `
select origin, is_deleted from storage.short_links where short=$1
`

// SqlSelectOriginAndShort select origin and short
const sqlSelectOriginAndShort = `
select origin, short from storage.short_links where user_id=$1
`

// sqlUpdate for set delete flag
const sqlUpdate = `
	UPDATE storage.short_links 
	SET is_deleted=true 
	WHERE user_id=$1 
	AND (correlation_id = ANY($2) OR short=ANY($3))
`

// New New new Storage with not null fields
func New(c *sql.DB, l *zap.Logger) (*PostgreSQLStorage, error) {
	// Check if scheme exist
	goose.SetBaseFS(migrations.EmbedMigrations)
	if err := goose.Up(c, "."); err != nil {
		panic(err)
	}

	return &PostgreSQLStorage{c, l}, nil
}

// LinkByShort implement interface for get data from storage by userId and shortLink
func (s *PostgreSQLStorage) LinkByShort(short shortlink.Short) (string, error) {
	var origin string
	var gone bool

	err := s.db.QueryRowContext(context.Background(), sqlSelectOrigin, string(short)).Scan(&origin, &gone)

	if err != nil {
		return "", er.ErrURLNotFound
	}

	if gone {
		return "", er.ErrURLIsGone
	}

	return origin, nil
}

// LinksByUser return all user links
func (s *PostgreSQLStorage) LinksByUser(userID user.UniqUser) (shortlink.ShortLinks, error) {
	origins := shortlink.ShortLinks{}
	rows, err := s.db.QueryContext(context.Background(), sqlSelectOriginAndShort, userID)
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
				return shortlink.Short(short), er.ErrAlreadyHasShort
			}
		}
		return short, err
	}
	return short, nil
}

// BunchSave save mass urls
func (s *PostgreSQLStorage) BunchSave(userID user.UniqUser, urls []shortlink.URLs) ([]shortlink.ShortURLs, error) {
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
		if _, err = stmt.ExecContext(context.Background(), userID, v.Origin, v.Short, v.ID); err == nil {
			shorts = append(shorts, shortlink.ShortURLs{
				Short: v.Short,
				ID:    v.ID,
			})
		} else {
			s.l.Info("Save bunch error", zap.Error(err))
		}
	}
	// Save changes
	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return shorts, nil
}

func (s *PostgreSQLStorage) Clear() error {
	if _, err := s.db.ExecContext(context.Background(), sqlDeleteRecords); err != nil {
		return err
	}
	return nil
}

// BunchUpdateAsDeleted  update as deleted
func (s *PostgreSQLStorage) BunchUpdateAsDeleted(ctx context.Context, ids []string, userID string) error {
	if len(ids) == 0 {
		return nil
	}
	idsArr := pq.Array(ids)
	_, err := s.db.ExecContext(ctx, sqlUpdate, userID, idsArr, idsArr)

	return err
}
