package evolution

import (
	"github.com/blang/semver"
	"github.com/markuszm/npm-analysis/model"
	"regexp"
	"strings"
	"time"
)

type MaintainerChange struct {
	PackageName string
	Name        string
	ReleaseTime time.Time
	ChangeType  string
	Version     string
}

func ProcessMaintainersSemVerSorted(metadata model.Metadata, timeCutoff time.Time) ([]MaintainerChange, error) {
	var changeList []MaintainerChange
	maintainersSet := make(map[string]bool)
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
		maintainers := ParseMaintainers(pkgData.Maintainers)
		seenMaintainers := make(map[string]bool, 0)
		releaseTime := GetTimeForVersion(metadata, vStr)

		if releaseTime.After(timeCutoff) {
			continue
		}

		for _, m := range maintainers {
			if !maintainersSet[m.Name] {
				changeType := "ADDED"
				if i == 0 {
					changeType = "INITIAL"
				}
				maintainersSet[m.Name] = true
				maintainerChange := MaintainerChange{
					PackageName: metadata.Name,
					Name:        m.Name,
					ReleaseTime: releaseTime,
					ChangeType:  changeType,
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
					ReleaseTime: releaseTime,
					ChangeType:  "REMOVED",
					Version:     vStr,
				}
				changeList = append(changeList, maintainerChange)
			}
		}
	}
	return changeList, nil
}

func ProcessMaintainersTimeSorted(metadata model.Metadata, timeCutoff time.Time) ([]MaintainerChange, error) {
	var changeList []MaintainerChange
	maintainersSet := make(map[string]bool)
	versions := metadata.Versions
	versionTimeList := SortTime(GetParsedTimeMap(metadata.Time))

	for _, v := range versionTimeList {
		vStr := v.Key
		pkgData := versions[vStr]
		if pkgData.Name == "" {
			continue
		}
		maintainers := ParseMaintainers(pkgData.Maintainers)
		seenMaintainers := make(map[string]bool, 0)

		releaseTime := GetTimeForVersion(metadata, vStr)
		if releaseTime.After(timeCutoff) {
			continue
		}

		for _, m := range maintainers {
			if !maintainersSet[m.Name] {
				changeType := "ADDED"
				if len(maintainersSet) == 0 {
					changeType = "INITIAL"
				}
				maintainerChange := MaintainerChange{
					PackageName: metadata.Name,
					Name:        m.Name,
					ReleaseTime: releaseTime,
					ChangeType:  changeType,
					Version:     vStr,
				}
				maintainersSet[m.Name] = true
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
					ReleaseTime: releaseTime,
					ChangeType:  "REMOVED",
					Version:     vStr,
				}
				changeList = append(changeList, maintainerChange)
			}
		}
	}
	return changeList, nil
}

func ParseMaintainers(maintainers interface{}) (persons []model.Person) {
	if maintainers == nil {
		return
	}
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
	return
}

func parsePersonObject(personObject interface{}) model.Person {
	if personObject == nil {
		return model.Person{}
	}

	var person model.Person

	switch personObject.(type) {
	case string:
		person = parseSingleMaintainerStr(personObject.(string))
	case map[string]interface{}:
		personMap := personObject.(map[string]interface{})
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

		person = model.Person{
			Name:  name.(string),
			Email: email.(string),
			Url:   url.(string),
		}
	}

	return person
}

const MaintainersRegex = `(^[a-zA-Z\s]+)`

func parseSingleMaintainerStr(str string) model.Person {
	r := regexp.MustCompile(MaintainersRegex)
	name := r.FindString(str)
	name = strings.TrimRight(name, " ")
	return model.Person{Name: name}
}
