package cmd

import (
	"fmt"
	"github.com/markuszm/npm-analysis/codeanalysis"
	"github.com/markuszm/npm-analysis/codeanalysispipeline"
	"github.com/spf13/cobra"
	"log"
)

var logPath string
var parallel bool
var maxWorkers int
var packagesPath string
var tmpPath string
var resultPath string
var collectorFlag string
var namesFilePath string
var loaderFlag string
var registryUrl string
var writerFlag string
var analysisFlag string

const mysqlUser = "root"
const mysqlPw = "npm-analysis"

// analysisCmd represents the analysis command
var analysisCmd = &cobra.Command{
	Use:   "analysis",
	Short: "Batch analysis of npm packages",
	Long:  `...`,
	Run: func(cmd *cobra.Command, args []string) {

		var collector codeanalysispipeline.NameCollector
		switch collectorFlag {
		case "db":
			logger.Infof("using db collector")
			var err error
			collector, err = codeanalysispipeline.NewDBNameCollector(fmt.Sprintf("%s:%s@/npm?charset=utf8mb4&collation=utf8mb4_bin", mysqlUser, mysqlPw))
			if err != nil {
				logger.Error(err)
			}
		case "file":
			logger.Infof("using file collector with file %v", namesFilePath)
			collector = codeanalysispipeline.NewFileNameCollector(namesFilePath)
		}

		//collector := codeanalysispipeline.NewTestNameCollector([]model.PackageVersionPair{
		//	{Name: "1720", Version: "1.0.0"},
		//	{Name: "@esm/ms", Version: "2.0.1"},
		//	{Name: "@fav/path", Version: "0.9.0"},
		//})

		var loader codeanalysispipeline.PackageLoader

		switch loaderFlag {
		case "disk":
			logger.Infof("using disk loader from packages path %v", packagesPath)
			loader = codeanalysispipeline.NewDiskLoader(packagesPath)
		case "net":
			logger.Infof("using net loader from registry %v and storing temp packages into path %v", registryUrl, tmpPath)
			loader = codeanalysispipeline.NewNetLoader(registryUrl, tmpPath)
		}

		unpacker := codeanalysispipeline.NewDiskUnpacker(tmpPath)

		var analysis codeanalysis.AnalysisExecutor
		switch analysisFlag {
		case "file_distribution":
			log.Print("executing file distribution analysis")
			analysis = codeanalysis.NewFileDistributionAnalysis(logger)
		case "used_dependencies":
			log.Print("executing used dependencies analysis")
			analysis = codeanalysis.NewUsedDependenciesAnalysis(logger)
		case "exports":
			log.Print("executing exports analysis")
			analysis = codeanalysis.NewExportsAnalysis(logger)
		}

		var writer codeanalysispipeline.ResultWriter
		switch writerFlag {
		case "csv":
			logger.Infof("using csv result writer to path %v", resultPath)
			writer = codeanalysispipeline.NewCSVWriter(resultPath)
		case "json":
			logger.Infof("using json result writer to path %v", resultPath)
			writer = codeanalysispipeline.NewJSONWriter(resultPath)
		}

		pipeline := codeanalysispipeline.NewPipeline(collector, loader, unpacker, analysis, writer, logger)

		var err error
		if parallel {
			logger.Infof("Running in parallel with %v workers", maxWorkers)
			err = pipeline.ExecuteParallel(maxWorkers)
		} else {
			logger.Info("Running in sequential mode")
			err = pipeline.Execute()
		}

		if err != nil {
			logger.Error(err)
		}

		// TODO on exit cleanup
	},
}

func init() {
	rootCmd.AddCommand(analysisCmd)

	analysisCmd.Flags().StringVarP(&logPath, "logfile", "l", "/home/markus/npm-analysis/codeanalysis.log", "path to log file")

	analysisCmd.Flags().BoolVarP(&parallel, "parallel", "m", false, "Execute pipeline in parallel?")
	analysisCmd.Flags().IntVarP(&maxWorkers, "scale", "s", 100, "number of workers (only if parallel)")

	analysisCmd.Flags().StringVarP(&packagesPath, "packages", "p", "/media/markus/NPM/NPM", "folder path to packages")
	analysisCmd.Flags().StringVarP(&tmpPath, "tmp", "t", "/home/markus/tmp", "Temp path to store extracted packages")
	analysisCmd.Flags().StringVarP(&resultPath, "output", "o", "/home/markus/npm-analysis/code-analysis.json", "File path to store results in")

	analysisCmd.Flags().StringVarP(&collectorFlag, "collector", "c", "db", "how to collect package names (db or file)")
	analysisCmd.Flags().StringVarP(&namesFilePath, "namesFile", "n", "./codeanalysispipeline/testfiles/test-packages.txt", "filepath containing package names")

	analysisCmd.Flags().StringVarP(&loaderFlag, "loader", "f", "disk", "specify loader type (disk or net)")
	analysisCmd.Flags().StringVarP(&registryUrl, "registry", "r", "http://registry.npmjs.org", "npm registry url (only when using net loader)")

	analysisCmd.Flags().StringVarP(&writerFlag, "writer", "w", "json", "specify writer type (csv or json)")

	analysisCmd.Flags().StringVarP(&analysisFlag, "analysis", "a", "file_distribution", "specify which analysis to run")
}
