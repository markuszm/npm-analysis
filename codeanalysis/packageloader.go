package codeanalysis

type PackageLoader interface {
	LoadPackage(packageName string) (string, error)
	LoadPackages(packageNames []string) (map[string]string, error)
}

type WebLoader struct {
	RegistryUrl string
	TempFolder  string
}

func NewWebLoader(registryUrl string, tempFolder string) *WebLoader {
	return &WebLoader{RegistryUrl: registryUrl, TempFolder: tempFolder}
}

func (w *WebLoader) LoadPackage(packageName string) (string, error) {
	return "", nil
}

func (w *WebLoader) LoadPackages(packageNames []string) (map[string]string, error) {
	result := make(map[string]string, len(packageNames))
	return result, nil
}

type DiskLoader struct {
	Path string
}

func NewDiskLoader(path string) (*DiskLoader, error) {
	return &DiskLoader{Path: path}, nil
}

func (d *DiskLoader) LoadPackage(packageName string) (string, error) {
	return "", nil
}

func (d *DiskLoader) LoadPackages(packageNames []string) (map[string]string, error) {
	result := make(map[string]string, len(packageNames))
	return result, nil
}
