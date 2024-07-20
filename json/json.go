package json

import (
	"encoding/json"
)

func parseJSON(data []byte) (map[string]any, error) {
	obj := map[string]any{}
	err := json.Unmarshal(data, &obj)
	return obj, err
}
