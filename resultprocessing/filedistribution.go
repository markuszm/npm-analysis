package resultprocessing

import (
	"encoding/csv"
	"encoding/json"
	"github.com/markuszm/npm-analysis/codeanalysispipeline"
	"github.com/markuszm/npm-analysis/util"
	"log"
	"os"
	"sort"
	"strconv"
)

func MergeFileDistributionResult(resultPath string, filter int) ([]util.Pair, error) {
	file, err := os.Open(resultPath)
	if err != nil {
		return nil, err
	}

	decoder := json.NewDecoder(file)

	mergedDistribution := make(map[string]int, 0)

	for {
		result := codeanalysispipeline.PackageResult{}
		err := decoder.Decode(&result)
		if err != nil {
			if err.Error() == "EOF" {
				log.Print("finished decoding result json")
				break
			} else {
				return nil, err
			}
		}

		switch result.Result.(type) {
		case map[string]interface{}:
			extensionMap := result.Result.(map[string]interface{})
			for ext, count := range extensionMap {
				// JSON numbers are decoded float64 so this transformation to int is necessary
				mergedDistribution[ext] += int(count.(float64))
			}
		default:
			continue
		}

	}

	var sortedDistribution []util.Pair

	for k, v := range mergedDistribution {
		if v > filter {
			sortedDistribution = append(sortedDistribution, util.Pair{Key: k, Value: v})
		}
	}

	sort.Sort(sort.Reverse(util.PairList(sortedDistribution)))

	return sortedDistribution, nil
}

func WriteFiledistributionResult(result []util.Pair, filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}

	writer := csv.NewWriter(file)
	defer writer.Flush()
	for _, r := range result {
		writer.Write([]string{r.Key, strconv.Itoa(r.Value)})
	}

	return nil
}

func CalculatePercentageForEachPackage(resultPath string) ([]PercentDistribution, error) {
	file, err := os.Open(resultPath)
	if err != nil {
		return nil, err
	}

	decoder := json.NewDecoder(file)

	var percentDistributions []PercentDistribution

	for {
		result := codeanalysispipeline.PackageResult{}
		err := decoder.Decode(&result)
		if err != nil {
			if err.Error() == "EOF" {
				log.Print("finished decoding result json")
				break
			} else {
				return nil, err
			}
		}

		totalFiles := 0.0

		percentageMap := make(map[string]float64, 0)

		switch result.Result.(type) {
		case map[string]interface{}:
			extensionMap := result.Result.(map[string]interface{})
			for ext, count := range extensionMap {
				// JSON numbers are decoded float64 so this transformation to int is necessary
				percentageMap[ext] = count.(float64)
				totalFiles += count.(float64)
			}
		default:
			continue
		}

		for ext, c := range percentageMap {
			percentage := c / totalFiles
			percentageMap[ext] = percentage
		}

		percentDistributions = append(percentDistributions, PercentDistribution{PackageName: result.Name, ExtensionMap: percentageMap})
	}
	return percentDistributions, nil
}

func WritePercentagesPerPackageForExtension(result []PercentDistribution, filePath string) error {
	distributionMap := make(map[string][]int)

	extensionToTrack := []string{
		"binary",
		".js",
		".json",
		".png",
		".md",
		".ts",
		"inary",
		".html",
		".xml",
		".map",
		".svg",
		".css",
		".npmignore",
		".scss",
		".class",
		".h",
		".txt",
		".yml",
		".less",
		".php",
		".hpp",
		".java",
		".gif",
		".rb",
		".c",
		".coffee",
		".jpg",
		".py",
		".jsx",
		".vue",
		".cpp",
		".d",
		".data",
		".woff",
		".sh",
		".ttf",
		".jshintrc",
		".cc",
		".m",
		".babelrc",
		".editorconfig",
		".eot",
		".flow",
		".lock",
		".eslintrc",
		".o",
		".hbs",
		".so",
		".plist",
		".dia",
		".jar",
		".dat",
		".ejs",
		".styl",
		".woff2",
		".glif",
		".csv",
		".jade",
		".tsx",
		".iml",
		".pcm",
		".timestamp",
		".gz",
		".ico",
		".properties",
		".as",
		".flat",
		".markdown",
		".dll",
		".gitattributes",
		".gitkeep",
		".diff",
		".eslintignore",
		".go",
		".pbf",
		".cs",
		".csl",
		".scssc",
		".tfm",
		".yaml",
		".tpl",
		".po",
		".opts",
		".buf",
		".gzip",
		".gyp",
		".tmpl",
		".pug",
		".htm",
		".sass",
		".pyc",
		".mo",
		".pem",
		".test",
		".tgz",
		".ml",
		".jst",
		".sql",
		".bat",
		".otf",
		".mustache",
		".bcmap",
		".gradle",
		".variables",
		".overrides",
		".vim",
		".nunjucks",
		".snap",
		".staticdata",
		".rst",
		".zip",
		".name",
		".strings",
		".es6",
		".log",
		".twig",
		".testcase",
		".hx",
		".proto",
		".cht",
		".sty",
		".ttl",
		".a",
		".crt",
		".exe",
		".pl",
		".pod",
		".conf",
		".pdf",
		".in",
		".glsl",
		".un~",
		".sjsinfo",
		".jscsrc",
		".pbxproj",
		".info",
		".out",
		".json5",
		".node",
		".bin",
		".handlebars",
		".flf",
		".mli",
		".tree",
		".swift",
		".gitignore",
		".mk",
		".tid",
		".sol",
		".def",
		".gitmodules",
		".mjs",
		".S",
		".flowconfig",
		".icns",
		".cmd",
		".vf",
		".webp",
		".fd",
		".geojson",
		".xcscheme",
		".bowerrc",
		".erb",
		".template",
		".mp3",
		".s",
		".patch",
		".sha1",
		".jshintignore",
		".phtml",
		".MD",
		".pp",
		".nvmrc",
		".meta",
		".jpeg",
		".pak",
		".sample",
		".cfg",
		".bud",
		".aidl",
		".keep",
		".nib",
		".gemspec",
		".ls",
		".ini",
		".xq",
		".wav",
		".asm",
		".mtx",
		".m4",
		".mm",
		".pde",
		".tex",
		".DS_Store",
		".res",
		".key",
		".elm",
		".am",
		".re",
		".jsm",
		".config",
		".js~",
		".project",
		".cmake",
		".snippets",
		".mf",
		".LinkFileList",
		".ipp",
		".pm",
		".tsv",
		".feature",
		".mml",
		".priv",
		".swf",
		".adoc",
		".gypi",
		".xcworkspacedata",
		".pom",
		".el",
		".pub",
		".sln",
		".inc",
		".tlog",
		".xsd",
		".pro",
		".pgm",
		".watchmanconfig",
		".repositories",
		".response",
		".hmap",
		".lua",
		".request",
		".xsl",
		".eps",
		".obj",
		".tcl",
		".BSD",
		".xlsx",
		".vcxproj",
		".err",
		".psd",
		".dex",
		".pb",
		".phpt",
		".fail",
		".ics",
		".default",
		".done",
		".rawproto",
		".md~",
		".haml",
		".vcproj",
		".docx",
		".jspm-hash",
		".qml",
		".text",
		".bmp",
		".dot",
		".Po",
		".example",
		".reference",
		".tscparams",
		".dart",
		".afm",
		".ogg",
		".lint",
		".pfb",
		".l",
		".sublime-project",
		".idx",
		".swig",
		".modulemap",
		".xib",
		".real",
		".xcconfig",
		".vm",
		".cshtml",
		".icloud",
		".symtab",
		".json~",
		".trans",
		".store",
		".dockerignore",
		".cmi",
		".xhtml",
		".t",
		".prettierrc",
		".p",
		".njk",
		".env",
		".mxml",
		".es",
		".tar",
		".MIT",
		".cache",
		".qmlc",
		".lnk",
		".pyo",
		".db",
		".xcactivitylog",
		".iced",
		".qbk",
		".jsonld",
		".xaml",
		".pdb",
		".spec",
		".hh",
		".ps1",
		".xlf",
		".bak",
		".JPG",
		".podspec",
		".module",
		".prefs",
		".kt",
		".exp",
		".htaccess",
		".tern-project",
		".dust",
		".jsbeautifyrc",
		".wxss",
		".xcuserstate",
		".beam",
		".settings",
		".webidl",
		".aar",
		".APACHE2",
		".make",
		".design",
		".purs",
		".csproj",
		".pegjs",
		".meteor-portable",
		".dist",
		".ittf",
		".dylib",
		".mcss",
		".marko",
		".ngdoc",
		".PNG",
		".cson",
		".pch",
		".mp4",
		".ROS",
		".types",
		".filters",
		".pass",
		".dimacs",
		".bundle",
		".lang",
		".inl",
		".sublime-workspace",
		".cnf",
		".enc",
		".elmo",
		".elmi",
		".cls",
		".tern-port",
		".targ",
		".dats",
		".val",
		".ext1",
		".swc",
		".cljs",
		".cl",
		".stub",
		".wxml",
		".plain",
		".TXT",
		".frag",
		".lisp",
		".liquid",
		".md5",
	}

	for _, ext := range extensionToTrack {
		distributionMap[ext] = make([]int, 11)
	}

	for _, p := range result {
		for _, ext := range extensionToTrack {
			percentageIndex := int(p.ExtensionMap[ext] * 10)
			distributionMap[ext][percentageIndex]++

		}
	}

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}

	writer := csv.NewWriter(file)
	defer writer.Flush()
	for ext, distribution := range distributionMap {
		asStrings := []string{ext}
		for _, d := range distribution {
			asStrings = append(asStrings, strconv.Itoa(d))
		}
		err := writer.Write(asStrings)
		if err != nil {
			return err
		}
	}

	return nil
}

type PercentDistribution struct {
	PackageName  string
	ExtensionMap map[string]float64
}
