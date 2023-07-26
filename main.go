package main

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"mysqldump/pkg/logging"
	"mysqldump/pkg/readconfig"
	"mysqldump/pkg/checkprocess"
	"mysqldump/pkg/mysqldump"

	_ "github.com/go-sql-driver/mysql"
)

// MainFunc: Checking for errors in the execution results
func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// Subfunc cleanupOldDumps: Time sort dump files 
func sortByModTime(files []os.FileInfo) {
        sort.Slice(files, func(i, j int) bool {
                return files[i].ModTime().Before(files[j].ModTime())
        })
}

// Mainfunc: Generation management of dump files
func cleanupOldDumps(config *readconfig.Config) error {
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
	logFile, err := logging.SetupLogger()

	if err != nil {
		log.Fatalf("%s systemd[1]: %s", time.Now().Format("Jan 02 15:04:05"), err)
	}
	defer logFile.Close()

	// Read config
	config, err := readconfig.ReadConfig()
	checkError(err)

	// MySQL Active check
	err = checkprocess.RunMySQLActiveCheck(config)
	checkError(err)

	// Running mysqldump
	err = mysqldump.RunMySQLDump(config)
	checkError(err)

	// Generation management of dump files
	err = cleanupOldDumps(config)
	checkError(err)
}
