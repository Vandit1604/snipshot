package models

import (
	"errors"
	"time"
)

var (
	ErrRecordNotFound      = errors.New("models: no matching record found")
	ErrDuplicateEmail      = errors.New("models: duplicate email")
	ErrInvalidCredenetials = errors.New("models: invalid credentials")
)

type Snippet struct {
	ID      int
	Title   string
	Content string
	Created time.Time
	Expires time.Time
}

type User struct {
	ID             int
	Name           string
	Email          string
	HashedPassword []byte
	Created        time.Time
}
