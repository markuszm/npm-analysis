package model

type Package struct {
	Name         string            `json:"name"`
	Version      string            `json:"version"`
	Description  string            `json:"description"`
	Keywords     []string          `json:"keywords"`
	Bugs         BugTracker        `json:"bugs"`
	Author       Person            `json:"author"`
	Contributors []Person          `json:"contributors"`
	Maintainers  []Person          `json:"maintainers"`
	Files        []string          `json:"files"`
	Man          []string          `json:"man"`
	Dependencies map[string]string `json:"dependencies"`
	Homepage     string            `json:"homepage"`
	Main         string            `json:"main"`
	NpmVersion   string            `json:"_npmVersion"`
	NodeVersion  string            `json:"_nodeVersion"`
}
