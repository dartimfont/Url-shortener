package service

import (
	"log/slog"
	"os"
	"strconv"
	"url-shortener/internal/lib/random"
)

type URLStorage interface {
	SaveURL(urlToSave string, alias string) (int64, error)
	GetURL(alias string) (string, error)
	DeleteURL(alias string) (int64, error)
	UpdateURL(alias string, newURL string) (int64, error)
}

type URLService struct {
	storage URLStorage
	log     *slog.Logger
}

var aliasLength, _ = strconv.Atoi(os.Getenv("ALIAS_LENGTH"))

func NewURLService(storage URLStorage, log *slog.Logger) *URLService {
	return &URLService{
		storage: storage,
		log:     log,
	}
}

func (s *URLService) SaveURL(urlToSave string, alias string) (int64, error) {
	if alias == "" {
		alias = random.NewRandomString(aliasLength)
	}
	return s.storage.SaveURL(urlToSave, alias)
}

func (s *URLService) GetURL(alias string) (string, error) {
	return s.storage.GetURL(alias)
}

func (s *URLService) DeleteURL(alias string) (int64, error) {
	return s.storage.DeleteURL(alias)
}

func (s *URLService) UpdateURL(alias string, newURL string) (int64, error) {
	return s.storage.UpdateURL(alias, newURL)
}
