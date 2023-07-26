package mysqldump

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"mysqldump/pkg/readconfig"
)

func RunMySQLDump(config *readconfig.Config) error {
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

