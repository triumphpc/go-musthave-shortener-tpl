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

// producer type for save links in file-storage
type producer struct {
	file   *os.File
	writer *bufio.Writer
}

// consumer get files from file storage
type consumer struct {
	file *os.File
	// заменяем reader на scanner
	scanner *bufio.Scanner
}

// newProducer instance new producer
func newProducer(fileName string) (*producer, error) {
	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE, 0777)
	if err != nil {
		return nil, err
	}

	return &producer{
		file:   file,
		writer: bufio.NewWriter(file),
	}, nil
}

// NewConsumer instance new consumer
func NewConsumer(fileName string) (*consumer, error) {
	f, err := os.OpenFile(fileName, os.O_RDONLY|os.O_CREATE, 0777)
	if err != nil {
		return nil, err
	}

	return &consumer{
		file: f,
	}, nil
}

// Close file for producer
func (p *producer) close() error {
	return p.file.Close()
}

// Close file for consumer
func (c *consumer) close() error {
	return c.file.Close()
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
	// Create new producer for write links to file storage
	p, err := newProducer(fs)
	if nil != err {
		return
	}

	// handle for file close
	defer func(p *producer) {
		err := p.close()
		if err != nil {
			panic(ErrFileStorageNotClose)
		}
	}(p)

	// Convert to gob
	ge := gob.NewEncoder(p.writer)
	// encode
	if err := ge.Encode(s.data); err != nil {
		return
	}
	err = p.writer.Flush()
	if nil != err {
		log.Println(err)
		return
	}
	return
}

// Load all links to map
func (s *FileStorage) Load() error {
	fs, err := configs.Instance().Param(configs.FileStoragePath)
	if err != nil || fs == configs.FileStoragePathDefault {
		return nil
	}
	c, err := NewConsumer(fs)
	if nil != err {
		return err
	}

	defer func(c *consumer) {
		err := c.close()
		if err != nil {
			panic(ErrFileStorageNotClose)
		}
	}(c)

	gd := gob.NewDecoder(c.file)
	if err := gd.Decode(&s.data); err != nil {
		if err != io.EOF {
			return err
		}
	}
	log.Println("Load links from file storage")
	return nil
}
