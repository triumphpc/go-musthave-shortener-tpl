package storage

import (
	"errors"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/helpers"
)

// ShortLink type for short link
type ShortLink string

// Storage it's global storage
type Storage struct {
	data map[ShortLink]string
}

// Repository interface for working with global repository
type Repository interface {
	LinkBy(sl ShortLink) (string, error)
	Save(url string) (sl ShortLink)
}

// ErrURLNotFound error by package level
var ErrURLNotFound = errors.New("url not found")

// LinkBy implement interface for get data from storage
func (s *Storage) LinkBy(sl ShortLink) (string, error) {
	l, ok := s.data[sl]
	if !ok {
		return l, ErrURLNotFound
	}
	return l, nil
}

// Save url in storage of short links
func (s *Storage) Save(url string) (sl ShortLink) {
	sl = ShortLink(helpers.RandomString(10))
	// Save in database
	s.data[sl] = url
	return
}

// New Instance new Storage with not null fields
func New() *Storage {
	return &Storage{
		data: make(map[ShortLink]string),
	}
}
