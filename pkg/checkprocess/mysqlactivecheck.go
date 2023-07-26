package checkprocess

import (
    "database/sql"
    "fmt"

    "mysqldump/pkg/readconfig"
    _ "github.com/go-sql-driver/mysql"
)

// Subfunc runMySQLActiveCheck : Create MySQL DSN(Data Source Name)
func createMySQLDSN(config *readconfig.Config) string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/", config.DBUsername, config.DBPassword, config.DBAddress, config.DBPort)
}

// Mainfunc: Check if it can connect to MySQL.
func RunMySQLActiveCheck(config *readconfig.Config) error {
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

