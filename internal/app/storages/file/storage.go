package file

import (
	"errors"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/configs"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/helpers"
	fw "github.com/triumphpc/go-musthave-shortener-tpl/internal/app/storages/file-wrapper"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/storages/memory"
)

var ErrFileStorageNotClose = errors.New("file storage did not close")

type Storage struct {
	data map[memory.ShortLink]string
}

// New Instance new Storage with not null fields
func New() (*Storage, error) {
	s := &Storage{
		data: make(map[memory.ShortLink]string),
	}
	err := s.Load()

	if err != nil {
		return s, err
	}
	return s, nil
}

// LinkBy implement interface for get data from storage
func (s *Storage) LinkBy(sl memory.ShortLink) (string, error) {
	l, ok := s.data[sl]
	if !ok {
		return l, memory.ErrURLNotFound
	}
	return l, nil
}

// Save url in storage of short links
func (s *Storage) Save(url string) (sl memory.ShortLink) {
	sl = memory.ShortLink(helpers.RandomString(10))
	// Save in map storage or rewrite current
	s.data[sl] = url
	fs, err := configs.Instance().Param(configs.FileStoragePath)
	if err != nil || fs == configs.FileStoragePathDefault {
		return
	}

	// Save to storage
	_ = fw.Write(fs, s.data)
	return
}

// Load all links to map
func (s *Storage) Load() error {
	fs, err := configs.Instance().Param(configs.FileStoragePath)
	if err != nil || fs == configs.FileStoragePathDefault {
		return nil
	}
	if err := fw.Read(fs, &s.data); err != nil {
		return err
	}
	return nil
}
