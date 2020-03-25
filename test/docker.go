package test

import (
	"bytes"
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/fortytw2/dockertest"
	"github.com/go-sql-driver/mysql"
)

type nopLogger struct{}

func (nop *nopLogger) Print(v ...interface{}) {}

// this is a hack instead of fixing the actual problem
var globalDockerDB *sql.DB
var dockertestOnce sync.Once

func setupDockerDBs(t func(v ...interface{}), seedFilePath string) (*sql.DB, func()) {
	dockertestOnce.Do(func() {
		_ = mysql.SetLogger(&nopLogger{})

		args := []string{
			"-e", "MYSQL_ALLOW_EMPTY_PASSWORD=T",
			"-e", "MYSQL_DATABASE=scrub_test",
		}
		args = append(args, "--tmpfs", "/var/lib/mysql:rw")

		container, err := dockertest.RunContainer("mysql:5.7.24", "3306", func(addr string) error {
			addr = strings.Replace(addr, "localhost", "127.0.0.1", -1)
			db, err := sql.Open("mysql", "root:@tcp("+addr+")/scrub_test?parseTime=true&charset=utf8mb4&loc=UTC")
			if err != nil {
				return err
			}

			return db.Ping()
		}, args...)
		if err != nil {
			t("could not start mysql, %s", err)
		}

		// mysqlism
		address := strings.Replace(container.Addr, "localhost", "127.0.0.1", -1)

		localDBConfig := &connectionConfig{
			Database: connectionArgs{
				User:     "root",
				Password: "",
				Name:     "jwfriese",
				Url:      address,
				TimeZone: "UTC",
			},
		}

		runSeed(seedFilePath, localDBConfig)
		db, err := sql.Open("mysql", localDBConfig.connectionUrl(true))
		if err != nil {
			log.Fatalf("Could not open database connection: %v", err)
		}

		db.SetMaxOpenConns(30)
		db.SetMaxIdleConns(5)

		globalDockerDB = db
	})

	return globalDockerDB, func() {}
}

func runSeed(filePath string, dbConfig *connectionConfig) {
	if filePath == "" {
		return
	}
	log.Println("running database seed from", filePath)

	sqlFile, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatal("could not run seed", err)
	}

	var cmd *exec.Cmd

	urlParts := strings.Split(dbConfig.Database.Url, ":")
	hostname := urlParts[0]
	port := urlParts[1]

	userArg := fmt.Sprintf("-u%v", dbConfig.Database.User)
	databaseNameArg := fmt.Sprintf("-D%v", dbConfig.Database.Name)
	portArg := fmt.Sprintf("-P%v", port)
	hostnameArg := fmt.Sprintf("-h%v", hostname)
	if dbConfig.Database.Password == "" {
		cmd = exec.Command("mysql", userArg, databaseNameArg, portArg, hostnameArg)
	} else {
		passwordArg := fmt.Sprintf("-p%v", dbConfig.Database.Password)
		cmd = exec.Command("mysql", userArg, databaseNameArg, portArg, hostnameArg, passwordArg)
	}

	errBuff := bytes.NewBuffer([]byte{})
	cmd.Stdout = os.Stdout
	cmd.Stderr = errBuff
	cmd.Stdin = bytes.NewReader(sqlFile)

	err = cmd.Start()
	if err != nil {
		cmdErrorMsg, _ := ioutil.ReadAll(errBuff)
		log.Fatalln("error running sql seed:", err, string(cmdErrorMsg))
	}

	err = cmd.Wait()
	if err != nil {
		cmdErrorMsg, _ := ioutil.ReadAll(errBuff)
		log.Fatalln("error running sql seed:", err, string(cmdErrorMsg))
	}
}
