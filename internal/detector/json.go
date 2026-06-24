package detector

import "encoding/json"

func jsonUnmarshal(data []byte, value any) error {
	return json.Unmarshal(data, value)
}
