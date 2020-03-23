package test

import (
	"database/sql"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"runtime/debug"
	"sync"
)

var closeDatabases func()

func CleanUpDocker() {
	if closeDatabases != nil {
		closeDatabases()
	}
}

func CreateDB(seedFile string) *sql.DB {
	if testDatabase != nil {
		return testDatabase
	}

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

	testDatabase = initializeTestDatabase(connectionConfig)
	if testDatabase == nil {
		log.Fatalf("could not initialize a test database")
	}

	err = applySeed(testDatabase, seedFile)
	if err != nil {
		log.Fatalf("could not seed the test database: %v", err)
	}

	return testDatabase
}

var (
	testDatabase *sql.DB
	dockerLock   sync.Mutex
)

const DefaultConfigFilepath = "test/config.yaml"

func initializeTestDatabase(config *connectionConfig) *sql.DB {
	if _, ok := os.LookupEnv("SHOW_TEST_LOGS"); !ok {
		log.SetOutput(createTestLogPipe())
	}

	if config.UseDocker {
		return testDatabase
	}

	var (
		db  *sql.DB
		err error
	)
	db, err = sql.Open("mysql", config.connectionUrl())
	if err != nil {
		db, err = sql.Open("mysql", config.mySQLConnectionUrl())
		if err != nil {
			debug.PrintStack()
			log.Fatalf("Could not connect to mysql: %v", err)
		}

		log.Println("[test] creating database")
		_, err = db.Exec(fmt.Sprintf(`CREATE DATABASE IF NOT EXISTS %s`, config.Database.Name))
	}

	if err != nil {
		log.Println(err)
		log.Fatalf("Could not create test database: %v", err)
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
