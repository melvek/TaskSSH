package action

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

func decodeWith[T any](raw any) (T, error) {
	var out T
	if raw == nil {
		return out, nil
	}
	switch v := raw.(type) {
	case T:
		return v, nil
	case map[string]any:
		b, _ := yaml.Marshal(v)
		if err := yaml.Unmarshal(b, &out); err != nil {
			return out, fmt.Errorf("invalid with parameters: %w", err)
		}
		return out, nil
	default:
		return out, fmt.Errorf("unsupported with type %T", raw)
	}
}
