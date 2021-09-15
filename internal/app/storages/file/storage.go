package file

import (
	"errors"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/configs"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/helpers"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/models/shortlink"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/models/user"
	fw "github.com/triumphpc/go-musthave-shortener-tpl/internal/app/storages/file-wrapper"
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
func (s *UserStorage) LinkByShort(userId user.UniqUser, short shortlink.Short) (string, error) {
	shorts, ok := s.data[userId]
	if !ok {
		return "", ErrURLNotFound
	}
	url, ok := shorts[short]
	if !ok {
		return "", ErrURLNotFound
	}
	return url, nil
}

// Save url in storage of short links
func (s *UserStorage) Save(userId user.UniqUser, url string) shortlink.Short {
	short := shortlink.Short(helpers.RandomString(10))
	// Get current urls for user
	currentUrls := shortlink.ShortLinks{}
	if urls, ok := s.data[userId]; ok {
		currentUrls = urls
	}
	currentUrls[short] = url
	// Save in map storage or rewrite current
	s.data[userId] = currentUrls
	// Save to file storage
	fs, err := configs.Instance().Param(configs.FileStoragePath)
	if err != nil || fs == configs.FileStoragePathDefault {
		return short
	}
	_ = fw.Write(fs, s.data)
	return short
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
