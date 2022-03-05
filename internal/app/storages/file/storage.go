// Package file contain methods for file storage
package file

import (
	"context"
	"errors"

	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/helpers"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/models/shortlink"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/models/user"
	fw "github.com/triumphpc/go-musthave-shortener-tpl/internal/app/storages/filewrapper"
)

// ErrURLNotFound error by package level
var ErrURLNotFound = errors.New("url not found")

// UserStorage for file storage
type UserStorage struct {
	data            map[user.UniqUser]shortlink.ShortLinks
	fileStoragePath string
}

// New Instance new Storage with not null fields
func New(fileStoragePath string) (*UserStorage, error) {
	s := &UserStorage{
		data:            make(map[user.UniqUser]shortlink.ShortLinks),
		fileStoragePath: fileStoragePath,
	}
	// Load from file storage
	if err := s.Load(); err != nil {
		return s, err
	}
	return s, nil
}

// LinkByShort implement interface for get data from storage by userId and shortLink
func (s *UserStorage) LinkByShort(short shortlink.Short) (string, error) {
	shorts, ok := s.data["all"]
	if !ok {
		return "", ErrURLNotFound
	}
	url, ok := shorts[short]
	if !ok {
		return "", ErrURLNotFound
	}
	return url, nil
}

// LinksByUser return all user links
func (s *UserStorage) LinksByUser(userID user.UniqUser) (shortlink.ShortLinks, error) {
	shorts, ok := s.data[userID]
	if !ok {
		return shorts, ErrURLNotFound
	}
	return shorts, nil
}

// Save url in storage of short links
func (s *UserStorage) Save(userID user.UniqUser, url string) (shortlink.Short, error) {
	short := shortlink.Short(helpers.RandomString(10))
	// Get current urls for user
	currentUrls := shortlink.ShortLinks{}
	if urls, ok := s.data[userID]; ok {
		currentUrls = urls
	}
	currentUrls[short] = url
	// Save in map storage or rewrite current
	s.data[userID] = currentUrls
	s.data["all"] = currentUrls
	// Save to file storage
	if s.fileStoragePath == "" {
		return short, nil
	}
	_ = fw.Write(s.fileStoragePath, s.data)
	return short, nil
}

// Load all links to map
func (s *UserStorage) Load() error {
	// If file storage not exists
	if s.fileStoragePath == "" {
		return nil
	}
	if err := fw.Read(s.fileStoragePath, &s.data); err != nil {
		return err
	}
	return nil
}

// BunchSave save mass urls
func (s *UserStorage) BunchSave(userID user.UniqUser, urls []shortlink.URLs) ([]shortlink.ShortURLs, error) {
	var shorts []shortlink.ShortURLs
	return shorts, nil
}

// Clear database
func (s *UserStorage) Clear() error {
	return nil
}

func (s *UserStorage) BunchUpdateAsDeleted(ctx context.Context, ids []string, userID string) error {
	return nil
}
