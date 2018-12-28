package flooding

import (
	"encoding/json"
)

func marshalToJSONStr(data interface{}) (string, error) {
	bytes, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func unmarshalFromJSONStr(data string, out interface{}) error {
	return json.Unmarshal([]byte(data), out)
}
