package evolution

import "github.com/markuszm/npm-analysis/model"

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
