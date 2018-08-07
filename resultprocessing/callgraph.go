package resultprocessing

type Call struct {
	FromFile     string   `json:"fromFile"`
	FromFunction string   `json:"fromFunction"`
	Receiver     string   `json:"receiver"`
	Module       []string `json:"modules"`
	ToFunction   string   `json:"toFunction"`
	Arguments    []string `json:"args"`
	IsLocal      bool     `json:"isLocal"`
}
