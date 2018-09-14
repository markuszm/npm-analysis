package packagecallgraph

import (
	"io/ioutil"
	"os"
	"path"
	"strings"
)

type WriteObject interface {
	GetFields() []string
}

type Package struct {
	name string // ID
}

func (p *Package) GetFields() []string { return []string{p.name, "Package"} }

type Module struct {
	name       string // ID
	moduleName string
}

func (m *Module) GetFields() []string {
	return []string{removeNewLines(m.name), removeNewLines(m.moduleName), "Module"}
}

type Class struct {
	name      string // ID
	className string
}

func (c *Class) GetFields() []string {
	return []string{removeNewLines(c.name), removeNewLines(c.className), "Class"}
}

type Function struct {
	name         string // ID
	functionName string
	functionType string
}

func (f *Function) GetFields() []string {
	return []string{removeNewLines(f.name), removeNewLines(f.functionName), f.functionType, "Function"}
}

type Relation struct {
	startID, endID, relType string
}

func (r *Relation) GetFields() []string {
	return []string{removeNewLines(r.startID), removeNewLines(r.endID), r.relType}
}

type CSVChannels struct {
	PackageChan  chan WriteObject
	ModuleChan   chan WriteObject
	ClassChan    chan WriteObject
	FunctionChan chan WriteObject
	RelationChan chan WriteObject
}

func removeNewLines(str string) string {
	return strings.Replace(str, "\n", "", -1)
}

func CreateHeaderFiles(folder string) error {
	headerPackages := "name:ID,:LABEL"
	headerModules := "name:ID,moduleName,:LABEL"
	headerClasses := "name:ID,className,:LABEL"
	headerFunctions := "name:ID,functionName,functionType,:LABEL"
	headerRelations := ":START_ID,:END_ID,:TYPE"

	err := ioutil.WriteFile(path.Join(folder, "packages-header.csv"), []byte(headerPackages), os.ModePerm)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(path.Join(folder, "modules-header.csv"), []byte(headerModules), os.ModePerm)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(path.Join(folder, "classes-header.csv"), []byte(headerClasses), os.ModePerm)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(path.Join(folder, "functions-header.csv"), []byte(headerFunctions), os.ModePerm)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(path.Join(folder, "relations-header.csv"), []byte(headerRelations), os.ModePerm)
	return err
}

// RELATIONSHIP LABELS
const callRelation = "CALL"
const containsFunction = "CONTAINS_FUNCTION"
const containsClass = "CONTAINS_CLASS"
const containsClassFunction = "CONTAINS_CLASS_FUNCTION"
const containsModule = "CONTAINS_MODULE"
const requiresPackage = "REQUIRES_PACKAGE"
const requiresModule = "REQUIRES_MODULE"
