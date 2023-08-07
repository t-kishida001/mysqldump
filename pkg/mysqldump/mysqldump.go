package mysqldump

import (
	"compress/gzip"
	"fmt"
	"io"
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
            "--defaults-extra-file=.sql.cnf",
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

        // compress dump file
        gzFileName := dumpFileName + ".gz"
        gzFile, err := os.Create(gzFileName)
        if err != nil {
            fmt.Println("圧縮用ファイルの作成に失敗しました")
            return err
        }
        defer gzFile.Close()

        gzipWriter := gzip.NewWriter(gzFile)
        defer gzipWriter.Close()

        dumpFile.Seek(0, 0) // Reset file pointer to the beginning of the dump file
        _, err = io.Copy(gzipWriter, dumpFile)
        if err != nil {
            fmt.Println("ファイルの圧縮に失敗しました")
            return err
        }

        // Delete the original dump file after compression
        err = os.Remove(dumpFileName)
        if err != nil {
            fmt.Println("元のダンプファイルの削除に失敗しました")
            return err
        }

    }
    return nil
}
