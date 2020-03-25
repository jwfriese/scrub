package test

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

type connectionArgs struct {
	Url      string `yaml:"url"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Name     string `yaml:"name"`
	TestName string `yaml:"testName"`
	Protocol string `yaml:"protocol"`
	TimeZone string `yaml:"timeZone"`
}

type connectionConfig struct {
	Database  connectionArgs `yaml:"database"`
	UseDocker bool           `yaml:"useDocker"`
}

func (config *connectionConfig) connectionUrl(connectDirectlyToDatabase bool) string {
	if connectDirectlyToDatabase {
		return fmt.Sprintf(
			"%s:%s@%s(%s)/%s?parseTime=true&multiStatements=true&charset=utf8mb4&loc=%s",
			config.Database.User,
			config.Database.Password,
			config.Database.Protocol,
			config.Database.Url,
			config.Database.Name,
			config.Database.TimeZone,
		)
	}

	return fmt.Sprintf(
		"%s:%s@%s(%s)/?parseTime=true&multiStatements=true&charset=utf8mb4&loc=%s",
		config.Database.User,
		config.Database.Password,
		config.Database.Protocol,
		config.Database.Url,
		config.Database.TimeZone,
	)
}

func loadConnectionConfig() (*connectionConfig, error) {
	testConfig := &connectionConfig{}
	err := loadConfigurationFromFile(testConfig)
	if err != nil {
		return nil, err
	}
	return testConfig, nil
}

func loadConfigurationFromFile(conf *connectionConfig) error {
	configFilePath := DefaultConfigFilepath
	if os.Getenv("TEST_CONFIG_FILE_PATH") != "" {
		configFilePath = os.Getenv("TEST_CONFIG_FILE_PATH")
	}
	pathToConfigFile := filepath.Join(
		os.Getenv("GOPATH"),
		"src",
		"github.com",
		"jwfriese",
		"scrub",
		configFilePath,
	)

	file, err := ioutil.ReadFile(pathToConfigFile)
	if err != nil {
		log.Println("no test configuration found, falling back to docker")
		conf.UseDocker = true
		return nil
	}

	err = yaml.Unmarshal(file, conf)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}
