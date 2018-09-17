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
	startID, endID string
}

func (r *Relation) GetFields() []string {
	return []string{removeNewLines(r.startID), removeNewLines(r.endID)}
}

type CSVChannels struct {
	PackageChan               chan WriteObject
	ModuleChan                chan WriteObject
	ClassChan                 chan WriteObject
	FunctionChan              chan WriteObject
	CallsChan                 chan WriteObject
	ContainsClassChan         chan WriteObject
	ContainsClassFunctionChan chan WriteObject
	ContainsFunctionChan      chan WriteObject
	ContainsModuleChan        chan WriteObject
	RequiresModuleChan        chan WriteObject
	RequiresPackageChan       chan WriteObject
}

func removeNewLines(str string) string {
	return strings.Replace(str, "\n", "", -1)
}

func CreateHeaderFiles(folder string) error {
	headerPackages := "name:ID(Package-ID),:LABEL"
	headerModules := "name:ID(Module-ID),moduleName,:LABEL"
	headerClasses := "name:ID(Class-ID),className,:LABEL"
	headerFunctions := "name:ID(Function-ID),functionName,functionType,:LABEL"
	headerCall := ":START_ID(Function-ID),:END_ID(Function-ID)"
	headerContainsClass := ":START_ID(Module-ID),:END_ID(Class-ID)"
	headerContainsClassFunction := ":START_ID(Class-ID),:END_ID(Function-ID)"
	headerContainsFunction := ":START_ID(Module-ID),:END_ID(Function-ID)"
	headerContainsModule := ":START_ID(Package-ID),:END_ID(Module-ID)"
	headerRequiresPackage := ":START_ID(Package-ID),:END_ID(Package-ID)"
	headerRequiresModule := ":START_ID(Module-ID),:END_ID(Module-ID)"

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

	err = ioutil.WriteFile(path.Join(folder, "calls-header.csv"), []byte(headerCall), os.ModePerm)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(path.Join(folder, "containsclass-header.csv"), []byte(headerContainsClass), os.ModePerm)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(path.Join(folder, "containsclassfunction-header.csv"), []byte(headerContainsClassFunction), os.ModePerm)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(path.Join(folder, "containsfunction-header.csv"), []byte(headerContainsFunction), os.ModePerm)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(path.Join(folder, "containsmodule-header.csv"), []byte(headerContainsModule), os.ModePerm)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(path.Join(folder, "requiresmodule-header.csv"), []byte(headerRequiresModule), os.ModePerm)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(path.Join(folder, "requirespackage-header.csv"), []byte(headerRequiresPackage), os.ModePerm)
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
