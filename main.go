package main

import (
	"database/sql"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
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

// SubFunc setupLogger: Create Log dir
func createLogDir() (string, error) {
	logDir := filepath.Join(".", "logs")
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		if err := os.Mkdir(logDir, 0755); err != nil {
			return "", err
		}
	}
	return logDir, nil
}

// SubFunc setupLogger: Open log file "mysqldump.log"
func openLogFile(logDir string) (*os.File, error) {
	logFilePath := filepath.Join(logDir, "mysqldump.log")
	logFile, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	return logFile, nil
}

// MainFunc: Setup Log file
func setupLogger() (*os.File, error) {
	logDir, err := createLogDir()
	if err != nil {
		return nil, fmt.Errorf("ログフォルダの作成に失敗しました: %s", err)
	}

	logFile, err := openLogFile(logDir)
	if err != nil {
		return nil, fmt.Errorf("ログファイルのオープンに失敗しました: %s", err)
	}

	// set up logging to both file and standard output.
	log.SetOutput(io.MultiWriter(logFile, os.Stdout))

	return logFile, nil
}

// MainFunc: Checking for errors in the execution results
func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// MainFunc: Reading config files.
func readConfig() (*Config, error) {
	// reading .env.txt
        content, err := ioutil.ReadFile(".env.txt") 
        if err != nil {
                return nil, err
        }

	// setting to config get .env.txt
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

        // reading .sql.cnf
        sqlCnfContent, err := ioutil.ReadFile(".sql.cnf")
        if err != nil {
                return nil, err
        }

	// setting to config get .sql.cnf
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

        // ff no directory is specified,current working directory will be used.
        if config.DumpDir == "" {
                wd, err := os.Getwd()
                if err != nil {
                        return nil, err
                }
                config.DumpDir = wd
        }

        return config, nil
}

// Subfunc runMySQLActiveCheck : Create MySQL DSN(Data Source Name)
func createMySQLDSN(config *Config) string {
        return fmt.Sprintf("%s:%s@tcp(%s:%s)/", config.DBUsername, config.DBPassword, config.DBAddress, config.DBPort)
}

// Mainfunc: Check if it can connect to MySQL.
func runMySQLActiveCheck(config *Config) error {
        dsn := createMySQLDSN(config) // DSNを生成
        db, err := sql.Open("mysql", dsn)
        if err != nil {
                fmt.Println("mysqlに接続できません")
                return err
        }
        defer db.Close()

        err = db.Ping()
        if err != nil {
                return err
        }
        return nil
}

// Mainfunc: Running mysqldump
func runMySQLDump(config *Config) error {
	for _, dbName := range config.Databases {
		dumpCmd := exec.Command(
			"mysqldump",
			"--defaults-extra-file=/home/bbix/prj/mysqldump/.sql.cnf",
			"--databases",
			dbName,
			"--single-transaction",
			"--set-gtid-purged=OFF",
		)

		// output dump file
		currentTime := time.Now().Format("2006-01-02_1504")
		dumpFileName := filepath.Join(config.DumpDir, fmt.Sprintf("dump_%s_%s.sql", dbName, currentTime))
		dumpFile, err := os.Create(dumpFileName)
		if err != nil {
			fmt.Println("dump用ファイルの作成に失敗しました")
			return err
		}
		defer dumpFile.Close()
		dumpCmd.Stdout = dumpFile

		// DEBUG: Print the mysqldump command executed by the program to standard output.
		//fmt.Println("mysqldumpコマンド:", dumpCmd.String())

		err = dumpCmd.Run()
		if err != nil {
			fmt.Println("mysqldumpの実行に失敗しました")
			return err
		}
	}
	return nil
}

// Subfunc cleanupOldDumps: Time sort dump files 
func sortByModTime(files []os.FileInfo) {
        sort.Slice(files, func(i, j int) bool {
                return files[i].ModTime().Before(files[j].ModTime())
        })
}

// Mainfunc: Generation management of dump files
func cleanupOldDumps(config *Config) error {
        files, err := ioutil.ReadDir(config.DumpDir)
        if err != nil {
                return err
        }

        dumpFilesMap := make(map[string][]os.FileInfo)

        for _, file := range files {
                if strings.HasPrefix(file.Name(), "dump_") && strings.HasSuffix(file.Name(), ".sql") {
                        dbName := strings.TrimSuffix(strings.TrimPrefix(file.Name(), "dump_"), ".sql")
                        dbName = strings.Split(dbName, "_")[0] // Remove the part after timestamp.
                        dumpFilesMap[dbName] = append(dumpFilesMap[dbName], file)
                }
        }

	// compare with the number specified in the config.
        for _, dbDumpFiles := range dumpFilesMap {
                if len(dbDumpFiles) <= config.DumpGenerations {
                        continue
                }

                sortByModTime(dbDumpFiles)

                // delete old file. 
                for i, f := range dbDumpFiles {
                        if i < len(dbDumpFiles)-config.DumpGenerations {
                                err = os.Remove(filepath.Join(config.DumpDir, f.Name()))
                                if err != nil {
                                        return err
                                }
                        }
                }
        }

        return nil
}

func main() {
	// Setup logfile
	logFile, err := setupLogger()
	if err != nil {
		log.Fatalf("%s systemd[1]: %s", time.Now().Format("Jan 02 15:04:05"), err)
	}
	defer logFile.Close()

	// Read config
	config, err := readConfig()
	checkError(err)

	// MySQL Active check
	err = runMySQLActiveCheck(config)
	checkError(err)

	// Running mysqldump
	err = runMySQLDump(config)
	checkError(err)

	// Generation management of dump files
	err = cleanupOldDumps(config)
	checkError(err)
}
