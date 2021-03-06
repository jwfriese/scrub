package scrub

import (
	"log"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jwfriese/scrub/test"
	"github.com/stretchr/testify/assert"
)

func TestScrubbingWorksWhenSpecifyingDatabase(t *testing.T) {
	log.SetFlags(log.LstdFlags | log.Llongfile)
	db := test.CreateDB("test/seed.sql", false)

	defer test.CleanUpDocker()

	subject := New(db)

	err := subject.Scrub(Config{
		Selectors: []Selector{{
			Database: "scrub_test",
			Table:    "items",
			Column:   "description",
			Wheres:   "",
		}},
		Method: String("example"),
	})

	assert.NoError(t, err)

	rows, err := db.Query(`
			SELECT id, description, price
			FROM scrub_test.items
		`)
	assert.NoError(t, err)

	defer rows.Close()

	for rows.Next() {
		var (
			nextID          int
			nextDescription string
			nextPrice       float64
		)

		err := rows.Scan(&nextID, &nextDescription, &nextPrice)
		assert.NoError(t, err)

		assert.Equal(t, "example", nextDescription)
	}
}

func TestScrubbingWorksWithoutSpecifyingDatabase(t *testing.T) {
	log.SetFlags(log.LstdFlags | log.Llongfile)
	db := test.CreateDB("test/seed.sql", true)

	defer test.CleanUpDocker()

	subject := New(db)

	err := subject.Scrub(Config{
		Selectors: []Selector{{
			Table:  "items",
			Column: "description",
			Wheres: "",
		}},
		Method: String("example"),
	})

	assert.NoError(t, err)

	rows, err := db.Query(`
			SELECT id, description, price
			FROM items
		`)
	assert.NoError(t, err)

	defer rows.Close()

	for rows.Next() {
		var (
			nextID          int
			nextDescription string
			nextPrice       float64
		)

		err := rows.Scan(&nextID, &nextDescription, &nextPrice)
		assert.NoError(t, err)

		assert.Equal(t, "example", nextDescription)
	}
}
