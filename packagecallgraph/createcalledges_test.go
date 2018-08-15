package packagecallgraph

import (
	"fmt"
	"github.com/markuszm/npm-analysis/database"
	"go.uber.org/zap"
	"testing"
)

func TestRequirePackageName(t *testing.T) {
	testCases := []struct {
		moduleName, expected string
	}{
		{"aws-sdk/clients/s3", "aws-sdk"},
		{"aws-sdk/clients", "aws-sdk"},
		{"aws-sdk", "aws-sdk"},
		{"@storybook/react", "@storybook/react"},
		{"@storybook/react/views", "@storybook/react"},
		{"@storybook/react/views/data", "@storybook/react"},
	}
	for _, test := range testCases {
		t.Run(fmt.Sprint(test), func(t *testing.T) {
			actual := getRequiredPackageName(test.moduleName)
			if actual != test.expected {
				t.Errorf("FAIL: Expected %v but got %v", test.expected, actual)
			}
		})
	}
}

// INTEGRATION TEST - NEEDS TO RUN WITH PACKAGE METADATA DB RUNNING
func TestModuleNameForPackageImport(t *testing.T) {
	testCases := []struct {
		moduleName, expected string
	}{
		{"aws-sdk/clients/s3", "clients/s3"},
		{"aws-sdk/clients", "clients"},
		{"aws-sdk", "lib/aws"},
		{"@storybook/react", "dist/client/index"},
		{"@storybook/react/views", "views"},
		{"@storybook/react/views/data", "views/data"},
		{"54a54r4gr56ea4rg654a64ag6e4", "index"},
		{"24game-solver", "index"},
		{"yrn_build", "lib/yrn_build"},
		{"rijs", "index"},
		{"1771278", "lib/1771278"},
		{"redux-nkvd", "index.min"},
	}
	var mysqlUrl = fmt.Sprintf("%s:%s@/npm?charset=utf8mb4&collation=utf8mb4_bin", "root", "npm-analysis")

	mysqlInitializer := &database.Mysql{}
	mysql, databaseInitErr := mysqlInitializer.InitDB(mysqlUrl)
	defer mysql.Close()
	if databaseInitErr != nil {
		t.Fatal(databaseInitErr)
	}

	callEdgeCreator := NewCallEdgeCreator("", "", 10, mysql, zap.NewNop().Sugar())
	for _, test := range testCases {
		t.Run(fmt.Sprint(test), func(t *testing.T) {
			actual := callEdgeCreator.getModuleNameForPackageImport(test.moduleName)
			if actual != test.expected {
				t.Errorf("FAIL: Expected %v but got %v", test.expected, actual)
			}
		})
	}
}
