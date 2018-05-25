package evolution

import (
	"fmt"
	"github.com/blang/semver"
	"github.com/markuszm/npm-analysis/model"
	"time"
)

type LicenseChange struct {
	PackageName, Version                 string
	LicenseFrom, LicenseTo, ChangeString string
	ReleaseTime                          time.Time
}

func ProcessLicenseChanges(metadata model.Metadata) ([]LicenseChange, error) {
	var changeList []LicenseChange
	previousLicense := ""
	versions := metadata.Versions
	var semvers semver.Versions
	for _, v := range versions {
		semverParsed := semver.MustParse(v.Version)
		semvers = append(semvers, semverParsed)
	}
	semver.Sort(semvers)

	for i, v := range semvers {
		vStr := v.String()
		pkgData := versions[vStr]
		license := ProcessLicense(pkgData)
		if license == "" {
			license = ProcessLicenses(pkgData)
		}
		if previousLicense != license {
			if i == 0 {
				previousLicense = license
				continue
			}
			maintainerChange := LicenseChange{
				PackageName:  metadata.Name,
				ReleaseTime:  GetTimeForVersion(metadata, vStr),
				LicenseFrom:  previousLicense,
				LicenseTo:    license,
				ChangeString: fmt.Sprintf("%v->%v", previousLicense, license),
				Version:      vStr,
			}
			changeList = append(changeList, maintainerChange)
			previousLicense = license
		}
	}
	return changeList, nil
}

func ProcessLicense(version model.PackageLegacy) string {
	licenseStr := ""
	license := version.License
	if license == nil {
		return licenseStr
	}
	switch license.(type) {
	case string:
		licenseStr = license.(string)
	case []interface{}:
		licenseList := license.([]interface{})
		for i, l := range licenseList {
			switch l.(type) {
			case map[string]interface{}:
				licenseMap := l.(map[string]interface{})
				licenseType := licenseMap["type"]
				if licenseType != nil {
					if len(licenseList) > 1 && i != len(licenseList)-1 {
						licenseStr += licenseType.(string) + "|"
					} else {
						licenseStr += licenseType.(string)
					}
				}
			case string:
				if len(licenseList) > 1 && i != len(licenseList)-1 {
					licenseStr += l.(string) + "|"
				} else {
					licenseStr += l.(string)
				}
			}
		}
	case map[string]interface{}:
		licenseMap := license.(map[string]interface{})
		licenseType := licenseMap["type"]
		if licenseType != nil {
			licenseStr = licenseType.(string)
		}
	}
	return licenseStr
}

func ProcessLicenses(version model.PackageLegacy) string {
	licenseStr := ""
	license := version.Licenses
	if license == nil {
		return licenseStr
	}
	switch license.(type) {
	case string:
		licenseStr = license.(string)
	case []interface{}:
		licenseList := license.([]interface{})
		for i, l := range licenseList {
			switch l.(type) {
			case map[string]interface{}:
				licenseMap := l.(map[string]interface{})
				licenseType := licenseMap["type"]
				if licenseType != nil {
					if len(licenseList) > 1 && i != len(licenseList)-1 {
						licenseStr += licenseType.(string) + "|"
					} else {
						// extra edge case because makemehapi made some error in their licenses in the first version
						switch licenseType.(type) {
						case string:
							licenseStr += licenseType.(string)
						case map[string]interface{}:
							licenseStr += licenseType.(map[string]interface{})["type"].(string)
						}
					}
				}
			case string:
				if len(licenseList) > 1 && i != len(licenseList)-1 {
					licenseStr += l.(string) + "|"
				} else {
					licenseStr += l.(string)
				}
			}
		}
	case map[string]interface{}:
		licenseMap := license.(map[string]interface{})
		licenseType := licenseMap["type"]
		if licenseType != nil {
			licenseStr = licenseType.(string)
		}
	}
	return licenseStr
}
