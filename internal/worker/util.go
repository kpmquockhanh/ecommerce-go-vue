package worker

import (
	"encoding/json"
	"fmt"
)

func decodePayload(payload interface{}, target interface{}) error {
	// Marshal then unmarshal to convert from generic JSON to typed struct
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}
	if err := json.Unmarshal(data, target); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}
	return nil
}
