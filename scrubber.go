package scrub

import (
	"database/sql"
	"fmt"
)

type Scrubber interface {
	Scrub(Config) error
}

func New(db *sql.DB) Scrubber {
	return &scrubber{
		db: db,
	}
}

type scrubber struct {
	db *sql.DB
}

func (s *scrubber) Scrub(c Config) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	for _, selector := range c.Selectors {
		if err = doScrub(tx, selector, c.Method); err != nil {
			rollbackErr := tx.Rollback()
			if rollbackErr != nil {
				return rollbackErr
			}
			return err
		}
	}

	return tx.Commit()
}

func doScrub(
	db *sql.Tx,
	selector Selector,
	method Method,
) error {
	query := "UPDATE"
	var fmts []interface{}
	if selector.Database == "" {
		query += " %s"
		fmts = append(fmts, selector.Table)
	} else {
		query += " %s.%s"
		fmts = append(fmts, selector.Database)
		fmts = append(fmts, selector.Table)
	}
	query += " SET %s = ?"
	fmts = append(fmts, selector.Column)
	if selector.Wheres != "" {
		query += " WHERE %s"
		fmts = append(fmts, selector.Wheres)
	}

	query = fmt.Sprintf(query, fmts...)

	value, err := method.Execute()
	if err != nil {
		return err
	}
	_, err = db.Exec(query, value)
	return err
}

func doScrubWithWheres(
	db *sql.Tx,
	selector Selector,
	method Method,
) error {
	query := fmt.Sprintf("UPDATE %s SET %s = ? WHERE %s", selector.Table, selector.Column, selector.Wheres)
	value, err := method.Execute()
	if err != nil {
		return err
	}
	_, err = db.Exec(query, value)
	return err
}

func doScrubWithoutWheres(
	db *sql.Tx,
	selector Selector,
	method Method,
) error {
	query := fmt.Sprintf("UPDATE %s SET %s = ?", selector.Table, selector.Column)
	value, err := method.Execute()
	if err != nil {
		return err
	}
	_, err = db.Exec(query, value)
	return err
}
