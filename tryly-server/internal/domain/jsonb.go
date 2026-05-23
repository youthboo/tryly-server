package domain

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type JSONB map[string]interface{}

func jsonBytesFromSQL(src interface{}, typeName string) ([]byte, error) {
	if src == nil {
		return nil, nil
	}
	switch v := src.(type) {
	case []byte:
		return v, nil
	case string:
		return []byte(v), nil
	default:
		return nil, fmt.Errorf("%s: unsupported type %T", typeName, src)
	}
}

func (j *JSONB) Scan(src interface{}) error {
	data, err := jsonBytesFromSQL(src, "JSONB")
	if err != nil {
		return err
	}
	if len(data) == 0 {
		*j = JSONB{}
		return nil
	}
	return json.Unmarshal(data, j)
}

func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(j)
}
