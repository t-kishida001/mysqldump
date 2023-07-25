package main

import (
	"database/sql"
	"fmt"
	//"io"
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
	DumpDir         string // ダンプファイルを保存するディレクトリを追加
}

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func readConfig() (*Config, error) {
	content, err := ioutil.ReadFile(".env.txt") // envファイルのパス
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
		case "DB_ADDRESS":
			config.DBAddress = value
		case "DB_USERNAME":
			config.DBUsername = value
		case "DB_PASSWORD":
			config.DBPassword = value
		case "DATABASES":
			config.Databases = strings.Split(value, ",")
		case "DUMP_GENERATIONS":
			gen, err := strconv.Atoi(value)
			if err != nil {
				return nil, err
			}
			config.DumpGenerations = gen
		case "DUMP_DIR":
			config.DumpDir = value // ダンプファイルを保存するディレクトリの指定を追加
		}
	}

	// ディレクトリが指定されていない場合は、実行ディレクトリを使用する
	if config.DumpDir == "" {
		wd, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		config.DumpDir = wd
	}

	return config, nil
}

func runMySQLActiveCheck(config *Config) error {
	// データベースに接続してみる
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s)/", config.DBUsername, config.DBPassword, config.DBAddress))
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

		currentTime := time.Now().Format("2006-01-02_1504") // "2006-01-02_1504"はGoの日付フォーマット

		dumpFileName := filepath.Join(config.DumpDir, fmt.Sprintf("dump_%s_%s.sql", dbName, currentTime))
		dumpFile, err := os.Create(dumpFileName)
		if err != nil {
			fmt.Println("dump用ファイルの作成に失敗しました")
			return err
		}
		defer dumpFile.Close()
		dumpCmd.Stdout = dumpFile

		// プログラムが実行するmysqldumpコマンドを標準出力に表示
		//fmt.Println("mysqldumpコマンド:", dumpCmd.String())

		err = dumpCmd.Run()
		if err != nil {
			fmt.Println("mysqldumpの実行に失敗しました")
			return err
		}
	}
	return nil
}

func cleanupOldDumps(config *Config) error {
        files, err := ioutil.ReadDir(config.DumpDir)
        if err != nil {
                return err
        }

        dumpFilesMap := make(map[string][]os.FileInfo)

        for _, file := range files {
                if strings.HasPrefix(file.Name(), "dump_") && strings.HasSuffix(file.Name(), ".sql") {
                        dbName := strings.TrimSuffix(strings.TrimPrefix(file.Name(), "dump_"), ".sql")
                        dbName = strings.Split(dbName, "_")[0] // タイムスタンプ以降の部分を取り除く
                        dumpFilesMap[dbName] = append(dumpFilesMap[dbName], file)
                }
        }

        for _, dbDumpFiles := range dumpFilesMap {
                if len(dbDumpFiles) <= config.DumpGenerations {
                        continue
                }

                // ファイルを更新日時でソート
                sortByModTime(dbDumpFiles)

                // 古いファイルを削除し、指定した世代数だけ最新のファイルを保持
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

//func cleanupOldDumps(config *Config) error {
//        files, err := ioutil.ReadDir(config.DumpDir)
//        if err != nil {
//                return err
//        }
//
//        dumpFiles := []os.FileInfo{}
//        for _, file := range files {
//                if strings.HasPrefix(file.Name(), "dump_") && strings.HasSuffix(file.Name(), ".sql") {
//                        dumpFiles = append(dumpFiles, file)
//                }
//        }
//
//        for _, file := range dumpFiles {
//                fileName := file.Name()
//                dbName := strings.TrimSuffix(strings.TrimPrefix(fileName, "dump_"), ".sql")
//		fmt.Printf("ファイル名: %s, データベース名: %s\n", fileName, dbName)
//                files, err := ioutil.ReadDir(config.DumpDir)
//                if err != nil {
//                        return err
//                }
//
//                // dbNameに対応するファイルのみを取得して世代管理を行う
//                dbDumpFiles := []os.FileInfo{}
//                for _, f := range files {
//                        if strings.HasPrefix(f.Name(), "dump_"+dbName) && strings.HasSuffix(f.Name(), ".sql") {
//                                dbDumpFiles = append(dbDumpFiles, f)
//                        }
//                }
//
//		fmt.Println(len(dbDumpFiles))   //test
//
//                if len(dbDumpFiles) <= config.DumpGenerations {
//                        continue
//                }
//
//                // 古い順に並べ替え
//                sortByModTime(dbDumpFiles)
//		//fmt.Println(dbDumpFiles)   // test
//
//                // 古いファイルを削除
//                for _, f := range dbDumpFiles[:len(dbDumpFiles)-config.DumpGenerations] {
//                        err = os.Remove(filepath.Join(config.DumpDir, f.Name()))
//                        if err != nil {
//                                return err
//                        }
//                }
//        }
//
//        return nil
//}


func sortByModTime(files []os.FileInfo) {
	sort.Slice(files, func(i, j int) bool {
		return files[i].ModTime().Before(files[j].ModTime())
	})
}



func main() {
	config, err := readConfig()
	checkError(err)

	err = runMySQLActiveCheck(config)
	checkError(err)

	err = runMySQLDump(config)
	checkError(err)

	err = cleanupOldDumps(config)
	checkError(err)
}

