package storage

import (
	"errors"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/helpers"
)

// ShortLink type for short link
type ShortLink string

// Storage it's global storage
type Storage map[ShortLink]string

// Repository interface for working with global repository
type Repository interface {
	LinkBy(sl ShortLink) (string, error)
	Save(url string) (sl ShortLink)
}

// LinkBy implement interface for get data from storage
func (s Storage) LinkBy(sl ShortLink) (string, error) {
	l, ok := s[sl]
	if !ok {
		return l, errors.New("url not found")
	}
	return l, nil
}

// Save url in storage of short links
func (s Storage) Save(url string) (sl ShortLink) {
	sl = ShortLink(helpers.RandomString(10))
	// Save in database
	s[sl] = url
	return
}
