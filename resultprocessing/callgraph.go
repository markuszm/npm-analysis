package resultprocessing

type Call struct {
	FromFile     string   `json:"fromFile"`
	FromFunction string   `json:"fromFunction"`
	Receiver     string   `json:"receiver"`
	Module       string   `json:"module"`
	ToFile       string   `json:"toFile"`
	ToFunction   string   `json:"toFunction"`
	Arguments    []string `json:"args"`
}
