package cmd

import (
	"encoding/json"
	"github.com/blang/semver"
	"github.com/markuszm/npm-analysis/database"
	"github.com/markuszm/npm-analysis/evolution"
	"github.com/markuszm/npm-analysis/util"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
	"path"
	"time"
)

var cveReachDBUrl string
var cveReachMongoUrl string
var cveReachCVEPath string
var cveReachResultPath string

var cveReachCmd = &cobra.Command{
	Use:   "cveReach",
	Short: "Aggregates package reach of cves",
	Long:  `...`,
	Run: func(cmd *cobra.Command, args []string) {
		initializeLogger()

		processVulns()
	},
}

func init() {
	rootCmd.AddCommand(cveReachCmd)

	cveReachCmd.Flags().StringVar(&cveReachDBUrl, "db", "root:npm-analysis@/npm?charset=utf8mb4&collation=utf8mb4_bin", "db url to evolution data")
	cveReachCmd.Flags().StringVar(&cveReachMongoUrl, "mongo", "mongodb://npm:npm123@localhost:27017", "mongodb url to evolution data")
	cveReachCmd.Flags().StringVar(&cveReachCVEPath, "cve", "./vulns", "path to vulns")
	cveReachCmd.Flags().StringVar(&cveReachResultPath, "output", "./output/node-vulns", "output path")
}

func processVulns() {
	infos, err := ioutil.ReadDir(cveReachCVEPath)
	if err != nil {
		logger.Fatal(err)
	}

	mysqlInitializer := &database.Mysql{}
	db, err := mysqlInitializer.InitDB(cveReachDBUrl)
	if err != nil {
		logger.Fatal(err)
	}
	defer db.Close()

	mongoDB := database.NewMongoDB(metadataMongoUrl, "npm", "packageReach")
	err = mongoDB.Connect()
	if err != nil {
		logger.Fatal(err)
	}
	defer mongoDB.Disconnect()

	var result []PatchedDetails

	var packages []string

	for _, i := range infos {
		if !i.IsDir() {
			continue
		}
		jsonPath := path.Join(cveReachCVEPath, i.Name(), "raw.json")
		contents, err := ioutil.ReadFile(jsonPath)
		if err != nil {
			logger.Fatal(err)
		}
		vuln := Vulnerability{}
		err = json.Unmarshal(contents, &vuln)
		if err != nil {
			logger.Fatal(err)
		}

		patchedRange, err := semver.ParseRange(vuln.PatchedVersions)
		if err != nil {
			logger.Errorw("could not parse semver range", "err", err, "package", vuln.ModuleName)
			continue
		}

		vulnRange, err := semver.ParseRange(vuln.VulnerableVersions)
		if err != nil {
			logger.Errorw("could not parse semver range", "err", err, "package", vuln.ModuleName)
			continue
		}

		versionChanges, err := database.GetVersionChangesForPackage(vuln.ModuleName, db)
		if err != nil {
			logger.Errorf("ERROR: retrieving version changes for package %v with %v", vuln.ModuleName, err)
		}

		evolution.SortVersionChange(versionChanges)

		var patchedVersions []evolution.VersionChange
		var vulnerableVersions []evolution.VersionChange

		for _, v := range versionChanges {
			version, err := semver.Parse(v.Version)
			if err != nil {
				logger.Errorf("cannot parse version %v for package %v", v.Version, v.PackageName)
				continue
			}

			if patchedRange(version) {
				patchedVersions = append(patchedVersions, v)
			}

			if vulnRange(version) {
				vulnerableVersions = append(vulnerableVersions, v)
			}
		}

		var vulnerableDate time.Time
		var patchedDate time.Time

		if len(vulnerableVersions) > 0 {
			vulnerableSince := vulnerableVersions[0]
			vulnerableDate = vulnerableSince.ReleaseTime
		} else {
			vulnerableDate = time.Date(2018, time.Month(3), 31, 0, 0, 0, 0, time.UTC)
		}

		if len(patchedVersions) > 0 {
			patchedSince := patchedVersions[0]
			patchedDate = patchedSince.ReleaseTime
		} else {
			patchedDate = time.Date(2018, time.Month(4), 2, 0, 0, 0, 0, time.UTC)
		}

		details := PatchedDetails{
			PackageName:    vuln.ModuleName,
			Id:             vuln.Id,
			VulnerableDate: vulnerableDate,
			PatchedDate:    patchedDate,
		}

		result = append(result, details)
		packages = append(packages, details.PackageName)
	}
	packageReachEvolution, activeCveEvolution, activePackagesLists, err := retrievePackageReach(result, mongoDB)

	bytes, err := json.Marshal(result)
	if err != nil {
		logger.Fatal(err)
	}

	resultFilePath := path.Join(cveReachResultPath, "patchTimings.json")
	err = ioutil.WriteFile(resultFilePath, bytes, os.ModePerm)
	if err != nil {
		logger.Fatal(err)
	}

	bytes, err = json.Marshal(packages)
	if err != nil {
		logger.Fatal(err)
	}

	resultFilePath = path.Join(cveReachResultPath, "cvePackages.json")
	err = ioutil.WriteFile(resultFilePath, bytes, os.ModePerm)
	if err != nil {
		logger.Fatal(err)
	}

	bytes, err = json.Marshal(packageReachEvolution)
	if err != nil {
		logger.Fatal(err)
	}

	resultFilePath = path.Join(cveReachResultPath, "reachEvolution.json")
	err = ioutil.WriteFile(resultFilePath, bytes, os.ModePerm)
	if err != nil {
		logger.Fatal(err)
	}

	bytes, err = json.Marshal(activeCveEvolution)
	if err != nil {
		logger.Fatal(err)
	}

	resultFilePath = path.Join(cveReachResultPath, "activeCveEvolution.json")
	err = ioutil.WriteFile(resultFilePath, bytes, os.ModePerm)
	if err != nil {
		logger.Fatal(err)
	}

	bytes, err = json.Marshal(activePackagesLists)
	if err != nil {
		logger.Fatal(err)
	}

	resultFilePath = path.Join(cveReachResultPath, "activePackagesLists.json")
	err = ioutil.WriteFile(resultFilePath, bytes, os.ModePerm)
	if err != nil {
		logger.Fatal(err)
	}
}

func retrievePackageReach(patchedDetails []PatchedDetails, mongoDB *database.MongoDB) ([]util.TimeValue, []util.TimeValue, map[time.Time][]PatchedDetails, error) {
	var cveReachPoints []util.TimeValue
	var activeCvePoints []util.TimeValue
	activeCvePackagesTimeMap := make(map[time.Time][]PatchedDetails, 0)
	for year := 2010; year <= 2018; year++ {
		startMonth := 1
		endMonth := 12
		if year == 2010 {
			startMonth = 11
		}
		if year == 2018 {
			endMonth = 4
		}
		for month := startMonth; month <= endMonth; month++ {
			date := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
			reachIntersection := make(map[string]bool, 0)
			var activeCvePackages []PatchedDetails
			activeCves := 0

			for _, p := range patchedDetails {
				if p.VulnerableDate.Before(date) && p.PatchedDate.After(date) {
					reachDocument, err := mongoDB.FindPackageReach(p.PackageName, date)
					if err != nil {
						continue
					}

					for _, r := range reachDocument.ReachedPackages {
						if !reachIntersection[r] {
							reachIntersection[r] = true
						}
					}

					activeCves++
					activeCvePackages = append(activeCvePackages, p)
				}

			}

			cveReachPoints = append(cveReachPoints, util.TimeValue{
				Key:   date,
				Value: float64(len(reachIntersection)),
			})

			activeCvePoints = append(activeCvePoints, util.TimeValue{
				Key:   date,
				Value: float64(activeCves),
			})

			activeCvePackagesTimeMap[date] = activeCvePackages
		}
	}
	return cveReachPoints, activeCvePoints, activeCvePackagesTimeMap, nil
}

type Vulnerability struct {
	Id                 int       `json:"id"`
	ModuleName         string    `json:"module_name"`
	PublishDate        time.Time `json:"publish_date"`
	PatchedVersions    string    `json:"patched_versions"`
	VulnerableVersions string    `json:"vulnerable_versions"`
}

type PatchedDetails struct {
	PackageName    string
	Id             int
	VulnerableDate time.Time
	PatchedDate    time.Time
}
