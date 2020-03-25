package test

import (
	"database/sql"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"sync"
)

var closeDatabases func()

func CleanUpDocker() {
	if closeDatabases != nil {
		closeDatabases()
	}
}

func CreateDB(seedFile string, connectDirectlyToDatabase bool) *sql.DB {
	var testDatabase *sql.DB
	connectionConfig, err := loadConnectionConfig()
	if err != nil {
		log.Fatalf("failed to load test db config: %v", err)
	}

	if connectionConfig.UseDocker {
		dockerLock.Lock()
		defer dockerLock.Unlock()
		db, cancel := setupDockerDBs(log.Fatal, "")
		closeDatabases = cancel
		testDatabase = db
		return db
	}

	if connectDirectlyToDatabase {
		testDatabase = initializeTestDatabase(connectionConfig, true)
		if testDatabase == nil {
			log.Fatalf("could not initialize a test database")
		}

		err = applySeed(testDatabase, seedFile)
		if err != nil {
			log.Fatalf("could not seed the test database: %v", err)
		}
	} else {
		testDatabase = initializeTestDatabase(connectionConfig, true)
		if testDatabase == nil {
			log.Fatalf("could not initialize a test database")
		}

		err = applySeed(testDatabase, seedFile)
		if err != nil {
			log.Fatalf("could not seed the test database: %v", err)
		}

		err = testDatabase.Close()
		if err != nil {
			log.Fatalf("could not close the database connection used to seed that DB: %v", err)
		}

		testDatabase = initializeTestDatabase(connectionConfig, false)
		if testDatabase == nil {
			log.Fatalf("could not initialize a test database")
		}
	}

	return testDatabase
}

var dockerLock sync.Mutex

const DefaultConfigFilepath = "test/config.yaml"

func initializeTestDatabase(config *connectionConfig, connectDirectlyToDatabase bool) *sql.DB {
	if _, ok := os.LookupEnv("SHOW_TEST_LOGS"); !ok {
		log.SetOutput(createTestLogPipe())
	}

	var (
		db  *sql.DB
		err error
	)
	db, err = sql.Open("mysql", config.connectionUrl(connectDirectlyToDatabase))
	if err != nil {
		log.Println(err)
		log.Fatalf("could not connect to mysql: %v", err)
	}

	db.SetMaxOpenConns(30)
	db.SetMaxIdleConns(5)

	return db
}

func createTestLogPipe() io.Writer {
	outFile, err := os.Create("test.log")
	if err != nil {
		log.Fatal(err)
	}

	return outFile
}

func applySeed(db *sql.DB, seedFilePath string) error {
	fileBytes, err := ioutil.ReadFile(seedFilePath)
	if err != nil {
		return fmt.Errorf("failed to read seed file: %v", err)
	}

	fileString := string(fileBytes)
	_, err = db.Exec(fileString)
	if err != nil {
		return fmt.Errorf("failed to apply seed to database: %v", err)
	}

	return nil
}
