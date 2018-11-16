package packagecallgraph

import (
	"database/sql"
	"fmt"
	"github.com/markuszm/npm-analysis/database"
	"log"
	"path/filepath"
	"strings"
)

func getRequiredPackageName(moduleName string) string {
	if strings.Contains(moduleName, "/") {
		parts := strings.Split(moduleName, "/")
		if strings.HasPrefix(moduleName, "@") {
			return fmt.Sprintf("%s/%s", parts[0], parts[1])
		}
		return parts[0]
	}
	return moduleName
}

func getMainModuleForPackage(mysqlDatabase *sql.DB, moduleName string) string {
	packageName := getRequiredPackageName(moduleName)
	if strings.Contains(moduleName, "/") && packageName != moduleName {
		moduleName := strings.Replace(moduleName, packageName+"/", "", -1)
		return moduleName
	}
	mainFile, err := database.MainFileForPackage(mysqlDatabase, packageName)
	if err != nil {
		log.Fatalf("error getting mainFile from database for moduleName %s with error %s", moduleName, err)
	}
	// cleanup main file
	mainFile = strings.TrimSuffix(mainFile, filepath.Ext(mainFile))
	mainFile = strings.TrimLeft(mainFile, "./")

	if mainFile == "" {
		return "index"
	}
	return mainFile
}
