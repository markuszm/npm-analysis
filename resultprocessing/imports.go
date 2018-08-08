package resultprocessing

type Import struct {
	Identifier string `json:"id"`
	FromModule string `json:"fromModule"`
	ModuleName string `json:"moduleName"`
	BundleType string `json:"bundleType"`
	Imported   string `json:"imported"`
}
