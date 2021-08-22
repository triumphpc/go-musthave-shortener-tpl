package storage

import (
	"errors"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/helpers"
)

// MockStorage Implement mock object for global storage
type MockStorage Storage

// GenerateMockData use fake data
func (s MockStorage) GenerateMockData() {

}

// LinkBy implement interface for get data from storage
func (s MockStorage) LinkBy(sl ShortLink) (string, error) {
	l, ok := s[sl]
	if !ok {
		return l, errors.New("url not found")
	}
	return l, nil
}

// Save url in storage of short links
func (s MockStorage) Save(url string) (sl ShortLink) {
	sl = ShortLink(helpers.RandomString(10))
	// Save in database
	s[sl] = url
	return
}
