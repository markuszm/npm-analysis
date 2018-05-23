package evolution

import (
	"github.com/markuszm/npm-analysis/model"
	"time"
)

func GetTimeForVersion(metadata model.Metadata, version string) time.Time {
	t := metadata.Time[version]

	return ParseTime(t)
}

func ParseTime(t interface{}) time.Time {
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

func GetParsedTimeMap(times map[string]interface{}) map[string]time.Time {
	r := make(map[string]time.Time, 0)
	for k, v := range times {
		r[k] = ParseTime(v)
	}
	return r
}
