package storage

import (
	"bufio"
	"encoding/gob"
	"errors"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/configs"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/helpers"
	"io"
	"os"
)

// ShortLink type for short link
type ShortLink string

// Storage it's global storage
type Storage struct {
	Data map[ShortLink]string
}

// Producer save links in file storage
type Producer interface {
	WriteEvent(event *Storage)
	Close() error
}

// Consumer read from file storage
type Consumer interface {
	ReadEvent() (*Storage, error)
	Close() error
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

// NewProducer instance new producer
func NewProducer(fileName string) (*producer, error) {
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
func (p *producer) Close() error {
	return p.file.Close()
}

// Close file for consumer
func (c *consumer) Close() error {
	return c.file.Close()
}

// Repository interface for working with global repository
type Repository interface {
	LinkBy(sl ShortLink) (string, error)
	Save(url string) (sl ShortLink)
	// Flush write all links to file-storage
	Flush() error
	// Load links from file storage
	Load(c configs.Config) error
}

// ErrURLNotFound error by package level
var ErrURLNotFound = errors.New("url not found")

// New Instance new Storage with not null fields
func New() *Storage {
	return &Storage{
		Data: make(map[ShortLink]string),
	}
}

// LinkBy implement interface for get data from storage
func (s *Storage) LinkBy(sl ShortLink) (string, error) {
	l, ok := s.Data[sl]
	if !ok {
		return l, ErrURLNotFound
	}
	return l, nil
}

// Save url in storage of short links
func (s *Storage) Save(url string) (sl ShortLink) {
	sl = ShortLink(helpers.RandomString(10))
	// Save in map storage or rewrite current
	s.Data[sl] = url
	return
}

// Flush all links to file storage
func (s *Storage) Flush() error {
	fileStoragePath, err := configs.Instance().Param(configs.FileStoragePath)
	if err != nil || fileStoragePath == configs.FileStoragePathDefault {
		return nil
	}

	// Create new producer for write links to file storage
	p, err := NewProducer(fileStoragePath)
	if nil != err {
		return err
	}
	// Convert to gob
	gobEncoder := gob.NewEncoder(p.writer)
	// encode
	if err := gobEncoder.Encode(s.Data); err != nil {
		return err
	}

	return p.writer.Flush()
}

// Load all links to map
func (s *Storage) Load(c configs.Config) error {
	fs, err := c.Param(configs.FileStoragePath)

	if err != nil || fs == configs.FileStoragePathDefault {
		return nil
	}
	cns, err := NewConsumer(fs)
	if nil != err {
		return err
	}

	gobDecoder := gob.NewDecoder(cns.file)
	if err := gobDecoder.Decode(&s.Data); err != nil {
		if err != io.EOF {
			return err
		}
	}
	return nil
}
