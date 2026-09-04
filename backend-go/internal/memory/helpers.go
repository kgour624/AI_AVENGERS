package memory

import "encoding/json"

// marshalJSON is a helper to serialize data to JSON bytes.
func marshalJSON(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}
