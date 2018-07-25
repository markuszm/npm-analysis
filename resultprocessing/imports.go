package resultprocessing

type Import struct {
	Identifier string `json:"id"`
	ModuleName string `json:"moduleName"`
	BundleType string `json:"bundleType"`
	Imported   string `json:"imported"`
}
