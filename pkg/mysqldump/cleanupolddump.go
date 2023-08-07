package mysqldump

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"mysqldump/pkg/readconfig"
)

// Subfunc sortByModTime: Time sort dump files
func sortByModTime(files []os.FileInfo) {
	sort.Slice(files, func(i, j int) bool {
		return files[i].ModTime().Before(files[j].ModTime())
	})
}

// Mainfunc CleanupOldDumps: Generation management of dump files
func CleanupOldDumps(config *readconfig.Config) error {
	files, err := ioutil.ReadDir(config.DumpDir)
	if err != nil {
		return err
	}

	dumpFilesMap := make(map[string][]os.FileInfo)

	for _, file := range files {
		if strings.HasPrefix(file.Name(), "dump_") && strings.HasSuffix(file.Name(), ".sql.gz") {
			dbName := strings.TrimSuffix(strings.TrimPrefix(file.Name(), "dump_"), ".sql.gz")
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
