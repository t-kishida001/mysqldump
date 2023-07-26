// pkg/logger/setuplogger.go
package logging

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"log"
)

// Create Log dir
func createLogDir() (string, error) {
	logDir := filepath.Join(".", "logs")
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		if err := os.Mkdir(logDir, 0755); err != nil {
			return "", err
		}
	}
	return logDir, nil
}

// Open log file "mysqldump.log"
func openLogFile(logDir string) (*os.File, error) {
	logFilePath := filepath.Join(logDir, "mysqldump.log")
	logFile, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	return logFile, nil
}

// Setup Initialize Log file
func SetupLogger() (*os.File, error) {
	logDir, err := createLogDir()
	if err != nil {
		return nil, fmt.Errorf("ログフォルダの作成に失敗しました: %s", err)
	}
	
	logFile, err := openLogFile(logDir)
	if err != nil {
		return nil, fmt.Errorf("ログファイルのオープンに失敗しました: %s", err)
	}
	
	// set log output to both file and standard output
	log.SetOutput(io.MultiWriter(logFile, os.Stdout))
	
	return logFile, nil
}

