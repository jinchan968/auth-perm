package model

import (
	"encoding/json"

	"auth-perm/internal/common/errors"
)

// JSON a flexible type for storing JSON data in GORM
type JSON map[string]interface{}

// Scan implements the sql.Scanner interface for the JSON type.
func (j *JSON) Scan(value interface{}) error {
	if value == nil {
		*j = make(JSON)
		return nil
	}

	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, j)
	case string:
		return json.Unmarshal([]byte(v), j)
	default:
		return errors.NewValidationError("无法将非字符串值扫描到JSON类型中")
	}
}

// Value implements the driver.Valuer interface for the JSON type.
func (j JSON) Value() (interface{}, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}
