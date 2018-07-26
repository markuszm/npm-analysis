package resultprocessing

type Export struct {
	ExportType string `json:"type"`
	Identifier string `json:"id"`
	BundleType string `json:"bundleType"`
	File       string `json:"file"`
	IsDefault  bool   `json:"isDefault"`
}
