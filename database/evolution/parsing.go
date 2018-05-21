package evolution

import (
	"github.com/blang/semver"
	"github.com/markuszm/npm-analysis/database/model"
	"regexp"
	"strings"
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

type MaintainerChange struct {
	PackageName string
	Name        string
	ReleaseTime time.Time
	ChangeType  string
	Version     string
}

func ProcessMaintainers(metadata model.Metadata) ([]MaintainerChange, error) {
	var changeList []MaintainerChange
	maintainersSet := make(map[string]bool)
	versions := metadata.Versions
	var semvers semver.Versions
	for _, v := range versions {
		semverParsed := semver.MustParse(v.Version)
		semvers = append(semvers, semverParsed)
	}
	semver.Sort(semvers)

	for _, v := range semvers {
		vStr := v.String()
		pkgData := versions[vStr]
		maintainers := ParseMaintainers(pkgData.Maintainers)
		seenMaintainers := make(map[string]bool, 0)
		for _, m := range maintainers {
			if !maintainersSet[m.Name] {
				maintainersSet[m.Name] = true
				maintainerChange := MaintainerChange{
					PackageName: metadata.Name,
					Name:        m.Name,
					ReleaseTime: ParseTime(metadata, vStr),
					ChangeType:  "ADDED",
					Version:     vStr,
				}
				changeList = append(changeList, maintainerChange)
				seenMaintainers[m.Name] = true
			} else {
				seenMaintainers[m.Name] = true
			}
		}

		for m, ok := range maintainersSet {
			if !ok {
				continue
			}
			if !seenMaintainers[m] {
				maintainersSet[m] = false
				maintainerChange := MaintainerChange{
					PackageName: metadata.Name,
					Name:        m,
					ReleaseTime: ParseTime(metadata, vStr),
					ChangeType:  "REMOVED",
					Version:     vStr,
				}
				changeList = append(changeList, maintainerChange)
			}
		}
	}
	return changeList, nil
}

const MaintainersRegex = `(^[a-zA-Z\s]+)`

func ParseMaintainers(maintainers interface{}) []model.Person {
	var persons []model.Person
	switch maintainers.(type) {
	case string:
		persons = append(persons, parseSingleMaintainerStr(maintainers.(string)))
	case []interface{}:
		for _, v := range maintainers.([]interface{}) {
			persons = append(persons, parsePersonObject(v))
		}
	case map[string]interface{}:
		persons = append(persons, parsePersonObject(maintainers))
	}
	return persons
}

func parsePersonObject(personObject interface{}) model.Person {
	personMap := personObject.(map[string]interface{})
	// todo: very dirty hack to avoid nil in map
	name := personMap["name"]
	if name == nil {
		name = ""
	}
	email := personMap["email"]
	if email == nil {
		email = ""
	}
	url := personMap["url"]
	if url == nil {
		url = ""
	}

	if name == "" {
		anotherPersonMap := personMap["0"]
		return parsePersonObject(anotherPersonMap)
	}

	person := model.Person{
		Name:  name.(string),
		Email: email.(string),
		Url:   url.(string),
	}
	return person
}

func parseSingleMaintainerStr(str string) model.Person {
	r := regexp.MustCompile(MaintainersRegex)
	name := r.FindString(str)
	name = strings.TrimRight(name, " ")
	return model.Person{Name: name}
}
