package resultprocessing

type Call struct {
	FromModule   string   `json:"fromModule"`
	FromFunction string   `json:"fromFunction"`
	Receiver     string   `json:"receiver"`
	Module       []string `json:"modules"`
	ToFunction   string   `json:"toFunction"`
	Arguments    []string `json:"args"`
	IsLocal      bool     `json:"isLocal"`
}
