package model

type Metadata struct {
	Name     string                   `json:"name"`
	Versions map[string]PackageLegacy `json:"versions"`
	Time     map[string]interface{}   `json:"time"`
}
