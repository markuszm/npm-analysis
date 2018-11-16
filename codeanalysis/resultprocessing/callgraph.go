package resultprocessing

import (
	"encoding/json"
	"github.com/pkg/errors"
	"reflect"
)

type Call struct {
	FromModule   string   `json:"fromModule"`
	FromFunction string   `json:"fromFunction"`
	Receiver     string   `json:"receiver"`
	ClassName    string   `json:"className"`
	Modules      []string `json:"modules"`
	ToFunction   string   `json:"toFunction"`
	Arguments    []string `json:"args"`
	IsLocal      bool     `json:"isLocal"`
}

func TransformToCalls(result interface{}) ([]Call, error) {
	if reflect.TypeOf(result).String() != "[]interface {}" {
		return nil, errors.New("Error parsing calls")
	}

	objs := result.([]interface{})

	var calls []Call

	for _, value := range objs {
		call := Call{}
		bytes, err := json.Marshal(value)
		if err != nil {
			return calls, err
		}
		err = json.Unmarshal(bytes, &call)
		if err != nil {
			return calls, err
		}
		calls = append(calls, call)
	}
	return calls, nil
}
