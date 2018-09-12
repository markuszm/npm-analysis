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
	License              interface{}            `json:"license"` // multiple possible values
	Licenses             interface{}            `json:"licenses"`
	Man                  interface{}            `json:"man"` // multiple possible values
	Dependencies         map[string]string      `json:"dependencies"`
	DevDependencies      map[string]string      `json:"devDependencies"`
	PeerDependencies     map[string]string      `json:"peerDependencies"`
	BundledDependencies  []string               `json:"bundledDependencies"`
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

// special type as package metadata is not validated at all from npm :/
type PackageLegacy struct {
	Name                 string                 `json:"name"`
	Version              string                 `json:"version"`
	Description          interface{}            `json:"description"`
	Keywords             interface{}            `json:"keywords"`
	Author               interface{}            `json:"author"`       // multiple possible values
	Contributors         interface{}            `json:"contributors"` // multiple possible values
	Maintainers          interface{}            `json:"maintainers"`
	Files                interface{}            `json:"files"`
	Man                  interface{}            `json:"man"` // multiple possible values
	DepsWrong            interface{}            `json:"Dependencies"`
	Dependencies         map[string]interface{} `json:"dependencies"`
	DevDependencies      map[string]interface{} `json:"devDependencies"`
	PeerDependencies     interface{}            `json:"peerDependencies"`
	BundledDependencies  interface{}            `json:"bundledDependencies"`
	OptionalDependencies interface{}            `json:"optionalDependencies"`
	License              interface{}            `json:"license"`
	Licenses             interface{}            `json:"licenses"`
	OS                   interface{}            `json:"os"`
	CPU                  interface{}            `json:"cpu"`
	Engines              interface{}            `json:"engines"`    // multiple possible values
	Scripts              interface{}            `json:"scripts"`    // multiple possible values
	Repository           interface{}            `json:"repository"` // multiple possible values
	Bugs                 interface{}            `json:"bugs"`       // multiple possible values
	PublishConfig        interface{}            `json:"publishConfig"`
	Homepage             interface{}            `json:"homepage"`
	Main                 interface{}            `json:"main"` // multiple possible values
	NpmVersion           interface{}            `json:"_npmVersion"`
	NodeVersion          string                 `json:"_nodeVersion"`
	Distribution         interface{}            `json:"dist"`
}

type PackageVersionPair struct {
	Name    string
	Version string
}

type PackageResult struct {
	Name    string
	Version string
	Result  interface{}
}
