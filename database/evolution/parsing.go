package evolution

import (
	"github.com/markuszm/npm-analysis/database/model"
	"time"
)

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

func ParseTime(metadata model.Metadata, version string) time.Time {
	t := metadata.Time[version]

	parsedTime := time.Unix(1, 0)

	if t == nil {
		return parsedTime
	}

	timeForVersion := ""

	switch t.(type) {
	case string:
		timeForVersion = t.(string)
	case map[string]interface{}:
		// unpublished
		pkgMap := t.(map[string]interface{})
		unpublishTime := pkgMap["time"]
		if unpublishTime != nil {
			timeForVersion = unpublishTime.(string)
		}
	}

	var err error
	if timeForVersion != "" {
		parsedTime, err = time.Parse(time.RFC3339, timeForVersion)
		if err != nil {
			parsedTime = time.Unix(1, 0)
		}

	}

	if parsedTime.Sub(time.Unix(0, 0)) == 0 {
		parsedTime = time.Unix(1, 0)
	}

	return parsedTime
}
