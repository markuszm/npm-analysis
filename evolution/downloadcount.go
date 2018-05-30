package evolution

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

const urlDateRange1 = "https://api.npmjs.org/downloads/range/2015-01-10:2015-04-11/"
const urlDateRange2 = "https://api.npmjs.org/downloads/range/2015-04-12:2016-10-12/"
const urlDateRange3 = "https://api.npmjs.org/downloads/range/2016-10-13:2018-04-13/"

const bulkDateRange1 = "https://api.npmjs.org/downloads/range/2015-01-10:2015-04-11/"
const bulkDateRange2 = "https://api.npmjs.org/downloads/range/2015-04-12:2016-04-11/"
const bulkDateRange3 = "https://api.npmjs.org/downloads/range/2016-04-12:2017-04-12/"
const bulkDateRange4 = "https://api.npmjs.org/downloads/range/2017-04-13:2018-04-13/"

func GetDownloadCountsForPackage(pkg string) (string, error) {
	pkgName := pkg
	if strings.Contains(pkg, "/") {
		pkgName = transformScopedName(pkg)
	}

	range1, err := getDownloadCounts(urlDateRange1, pkgName)
	if err != nil {
		return "", err
	}
	range2, err := getDownloadCounts(urlDateRange2, pkgName)
	if err != nil {
		return "", err
	}
	range3, err := getDownloadCounts(urlDateRange3, pkgName)
	if err != nil {
		return "", err
	}

	fullRangeDownloadCounts := appendAllSlices(range1.DownloadCounts, range2.DownloadCounts, range3.DownloadCounts)

	fullRange := DownloadCountResponse{
		Start:          range1.Start,
		End:            range3.End,
		Package:        pkgName,
		DownloadCounts: fullRangeDownloadCounts,
	}

	bytes, err := json.Marshal(fullRange)
	if err != nil {
		return "", err
	}

	doc := string(bytes)
	return doc, err
}

func GetDownloadCountsBulk(packages []string) (map[string]string, error) {
	allDownloadCounts := make(map[string]string)

	range1, err := getDownloadCountsBulk(bulkDateRange1, packages)
	if err != nil {
		return allDownloadCounts, err
	}
	range2, err := getDownloadCountsBulk(bulkDateRange2, packages)
	if err != nil {
		return allDownloadCounts, err
	}
	range3, err := getDownloadCountsBulk(bulkDateRange3, packages)
	if err != nil {
		return allDownloadCounts, err
	}
	range4, err := getDownloadCountsBulk(bulkDateRange4, packages)
	if err != nil {
		return allDownloadCounts, err
	}

	for _, p := range packages {
		fullRangeDownloadCounts := appendAllSlices(range1[p].DownloadCounts, range2[p].DownloadCounts, range3[p].DownloadCounts, range4[p].DownloadCounts)

		fullRange := DownloadCountResponse{
			Start:          range1[p].Start,
			End:            range4[p].End,
			Package:        p,
			DownloadCounts: fullRangeDownloadCounts,
		}

		bytes, err := json.Marshal(fullRange)
		if err != nil {
			return allDownloadCounts, err
		}

		doc := string(bytes)
		allDownloadCounts[p] = doc
	}
	return allDownloadCounts, nil
}

func appendAllSlices(slices ...[]DownloadCountPerDay) []DownloadCountPerDay {
	// Source: https://stackoverflow.com/questions/37884361/concat-multiple-slices-in-golang
	var totalLen int
	for _, s := range slices {
		totalLen += len(s)
	}
	tmp := make([]DownloadCountPerDay, totalLen)
	var i int
	for _, s := range slices {
		i += copy(tmp[i:], s)
	}
	return tmp
}

func getDownloadCounts(url, pkg string) (DownloadCountResponse, error) {
	downloadCounts := DownloadCountResponse{}

	fullUrl := url + pkg

	resp, err := http.Get(fullUrl)
	if err != nil {
		return downloadCounts, err
	}
	defer resp.Body.Close()

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return downloadCounts, err
	}

	json.Unmarshal(bytes, &downloadCounts)
	return downloadCounts, nil
}

func getDownloadCountsBulk(url string, pkg []string) (map[string]DownloadCountResponse, error) {
	downloadCountResponses := make(map[string]DownloadCountResponse, 0)

	fullUrl := url
	for _, p := range pkg {
		if strings.HasPrefix(p, "@") {
			continue
		}
		fullUrl += p + ","
	}

	fullUrl = strings.TrimRight(fullUrl, ",")

	resp, err := http.Get(fullUrl)
	if err != nil {
		return downloadCountResponses, err
	}
	defer resp.Body.Close()

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return downloadCountResponses, err
	}

	json.Unmarshal(bytes, &downloadCountResponses)
	return downloadCountResponses, nil
}

func MustParseDate(str string) time.Time {
	date, err := time.Parse("2006-01-02", str)
	if err != nil {
		log.Fatalf("ERROR: Could not parse date with error: %v", err)
	}
	return date
}

type DownloadCountResponse struct {
	Start          string                `json:"start"`
	End            string                `json:"end"`
	Package        string                `json:"package"`
	DownloadCounts []DownloadCountPerDay `json:"downloads"`
}

type DownloadCountPerDay struct {
	Downloads int    `json:"downloads"`
	Day       string `json:"day"`
}
