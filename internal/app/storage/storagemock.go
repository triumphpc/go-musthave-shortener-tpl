package storage

import (
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/helpers"
	"strconv"
)

// MockStorage Implement mock object for global storage
type MockStorage Storage

// Number of tests
var testCount = 0

// GenerateMockData use fake data
func (s *MockStorage) GenerateMockData() {
	s.Save("http://yandex.ru")
	s.Save("http://google.ru")
	s.Save("http://localhost")
}

// LinkBy implement interface for get data from storage
func (s *MockStorage) LinkBy(sl ShortLink) (string, error) {
	if s.data == nil {

		return "", ErrUrlNotFound
	}

	l, ok := s.data[sl]
	if !ok {
		return l, ErrUrlNotFound
	}
	return l, nil
}

// Save url in storage of short links
func (s *MockStorage) Save(url string) (sl ShortLink) {
	// Generate tests links
	testCount = testCount + 1
	sl = ShortLink(helpers.RandomString(10) + "_test_" + strconv.Itoa(testCount))

	if s.data == nil {
		s.data = make(map[ShortLink]string)
	}
	// Save in database
	s.data[sl] = url
	return
}
