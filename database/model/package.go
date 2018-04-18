package model

type Package struct {
	Name                 string                 `json:"name"`
	Version              string                 `json:"version"`
	Description          string                 `json:"description"`
	Keywords             []string               `json:"keywords"`
	Author               interface{}            `json:"author"`       // multiple possible values
	Contributors         interface{}            `json:"contributors"` // multiple possible values
	Maintainers          []Person               `json:"maintainers"`
	Files                []string               `json:"files"`
	Man                  interface{}            `json:"man"` // multiple possible values
	Dependencies         map[string]string      `json:"dependencies"`
	DevDependencies      map[string]string      `json:"devDependencies"`
	PeerDependencies     map[string]string      `json:"peerDependencies"`
	BundledDependencies  map[string]string      `json:"bundledDependencies"`
	OptionalDependencies map[string]string      `json:"optionalDependencies"`
	OS                   []string               `json:"os"`
	CPU                  []string               `json:"cpu"`
	Engines              interface{}            `json:"engines"`    // multiple possible values
	Scripts              interface{}            `json:"scripts"`    // multiple possible values
	Repository           interface{}            `json:"repository"` // multiple possible values
	Bugs                 interface{}            `json:"bugs"`       // multiple possible values
	PublishConfig        map[string]interface{} `json:"publishConfig"`
	Homepage             string                 `json:"homepage"`
	Main                 interface{}            `json:"main"` // multiple possible values
	NpmVersion           string                 `json:"_npmVersion"`
	NodeVersion          string                 `json:"_nodeVersion"`
	Distribution         Dist                   `json:"dist"`
}
