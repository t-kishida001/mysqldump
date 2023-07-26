package main

import (
	"log"
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
	err = mysqldump.CleanupOldDumps(config)
	checkError(err)
}
