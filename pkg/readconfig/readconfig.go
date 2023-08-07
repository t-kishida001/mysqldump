package readconfig

import (
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	DBAddress       string
	DBUsername      string
	DBPassword      string
	Databases       []string
	DumpGenerations int
	DumpDir         string
	DBPort          string
}

// ReadConfig reads the config files.
func ReadConfig() (*Config, error) {
	content, err := ioutil.ReadFile(".env.txt")
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(content), "\n")
	config := &Config{}
	for _, line := range lines {
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "DATABASES":
			config.Databases = strings.Split(value, ",")
		case "DUMP_GENERATIONS":
			gen, err := strconv.Atoi(value)
			if err != nil {
				return nil, err
			}
			config.DumpGenerations = gen
		case "DUMP_DIR":
			config.DumpDir = value
		}
	}

	sqlCnfContent, err := ioutil.ReadFile(".sql.cnf")
	if err != nil {
		return nil, err
	}

	lines = strings.Split(string(sqlCnfContent), "\n")
	for _, line := range lines {
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "user":
			config.DBUsername = value
		case "password":
			config.DBPassword = value
		case "host":
			config.DBAddress = value
		case "port":
			config.DBPort = value
		}
	}

	if config.DumpDir == "" {
		wd, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		config.DumpDir = wd
	}

	return config, nil
}
