package scrub

import (
	"database/sql/driver"

	"github.com/brianvoe/gofakeit/v4"
)

type Method interface {
	Execute() (driver.Value, error)
}

type SetString struct {
	Value string
}

func (m *SetString) Execute() (driver.Value, error) {
	return m.Value, nil
}

func RandomWords(count int) *SetString {
	return &SetString{
		Value: gofakeit.Sentence(count),
	}
}

func String(v string) *SetString {
	return &SetString{
		Value: v,
	}
}
