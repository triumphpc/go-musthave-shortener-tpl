package file

import (
	"errors"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/configs"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/helpers"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/models/shortlink"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/models/user"
	fw "github.com/triumphpc/go-musthave-shortener-tpl/internal/app/storages/filewrapper"
)

// ErrURLNotFound error by package level
var ErrURLNotFound = errors.New("url not found")

// UserStorage for file storage
type UserStorage struct {
	data map[user.UniqUser]shortlink.ShortLinks
}

// New Instance new Storage with not null fields
func New() (*UserStorage, error) {
	s := &UserStorage{
		data: make(map[user.UniqUser]shortlink.ShortLinks),
	}
	// Load from file storage
	if err := s.Load(); err != nil {
		return s, err
	}
	return s, nil
}

// LinkByShort implement interface for get data from storage by userId and shortLink
func (s *UserStorage) LinkByShort(short shortlink.Short, userID user.UniqUser) (string, error) {
	shorts, ok := s.data[userID]
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
	fs, err := configs.Instance().Param(configs.FileStoragePath)
	if err != nil || fs == configs.FileStoragePathDefault {
		return short, nil
	}
	_ = fw.Write(fs, s.data)
	return short, nil
}

// Load all links to map
func (s *UserStorage) Load() error {
	fs, err := configs.Instance().Param(configs.FileStoragePath)
	// If file storage not exists
	if err != nil || fs == configs.FileStoragePathDefault {
		return nil
	}
	if err := fw.Read(fs, &s.data); err != nil {
		return err
	}
	return nil
}

// BunchSave save mass urls
func (s *UserStorage) BunchSave(urls []shortlink.URLs, userID user.UniqUser) ([]shortlink.ShortURLs, error) {
	var shorts []shortlink.ShortURLs
	return shorts, nil
}
