package filestorage

import (
	"bufio"
	"encoding/gob"
	"errors"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/configs"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/helpers"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/storage"
	"io"
	"log"
	"os"
)

var ErrFileStorageNotClose = errors.New("file storage did not close")

type FileStorage struct {
	data map[storage.ShortLink]string
}

// New Instance new Storage with not null fields
func New() (*FileStorage, error) {
	s := &FileStorage{
		data: make(map[storage.ShortLink]string),
	}
	err := s.Load()

	if err != nil {
		return s, err
	}
	return s, nil
}

// LinkBy implement interface for get data from storage
func (s *FileStorage) LinkBy(sl storage.ShortLink) (string, error) {
	l, ok := s.data[sl]
	if !ok {
		return l, storage.ErrURLNotFound
	}
	return l, nil
}

// Save url in storage of short links
func (s *FileStorage) Save(url string) (sl storage.ShortLink) {
	sl = storage.ShortLink(helpers.RandomString(10))
	// Save in map storage or rewrite current
	s.data[sl] = url
	fs, err := configs.Instance().Param(configs.FileStoragePath)
	if err != nil || fs == configs.FileStoragePathDefault {
		return
	}

	file, err := os.OpenFile(fs, os.O_WRONLY|os.O_CREATE, 0777)
	if err != nil {
		return
	}

	// handle for file close
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			panic(ErrFileStorageNotClose)
		}
	}(file)
	// Convert to gob
	buffer := bufio.NewWriter(file)
	ge := gob.NewEncoder(buffer)
	// encode
	if err := ge.Encode(s.data); err != nil {
		return
	}
	_ = buffer.Flush()
	return
}

// Load all links to map
func (s *FileStorage) Load() error {
	fs, err := configs.Instance().Param(configs.FileStoragePath)
	if err != nil || fs == configs.FileStoragePathDefault {
		return nil
	}

	file, err := os.OpenFile(fs, os.O_RDONLY|os.O_CREATE, 0777)
	if err != nil {
		return err
	}

	// handle for file close
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			panic(ErrFileStorageNotClose)
		}
	}(file)

	gd := gob.NewDecoder(file)
	if err := gd.Decode(&s.data); err != nil {
		if err != io.EOF {
			return err
		}
	}
	log.Println("Load links from file storage")
	return nil
}
