package resultprocessing

type Export struct {
	ExportType string   `json:"type"`
	Identifier string   `json:"id"`
	Arguments  []string `json:"args"`
	BundleType string   `json:"bundleType"`
	File       string   `json:"file"`
	IsDefault  bool     `json:"isDefault"`
	Local      string   `json:"local"`
}
