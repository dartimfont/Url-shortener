package domain

import "errors"

type URL struct {
	ID     string
	Alias  string
	NewURL string
}

var (
	ErrURLNotFound = errors.New("url not found")
	ErrURLExists   = errors.New("url exists")
)
