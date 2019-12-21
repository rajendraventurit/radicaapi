package db

import (
	"database/sql"
	"encoding/json"
)

// NullBool wraps a sql.NullBool to allow proper json conversion
type NullBool struct {
	sql.NullBool
}

// MarshalJSON will convert NullBool to json value or null
func (v NullBool) MarshalJSON() ([]byte, error) {
	if v.Valid {
		return json.Marshal(v.Bool)
	}
	return json.Marshal(nil)
}

// UnmarshalJSON will return json encoded for value
func (v *NullBool) UnmarshalJSON(data []byte) error {
	// Unmarshalling into a pointer will let us detect null
	var x *bool
	if err := json.Unmarshal(data, &x); err != nil {
		return err
	}
	if x != nil {
		v.Valid = true
		v.Bool = *x
	} else {
		v.Valid = false
	}
	return nil
}

// NewNullBool returns a valid NullBool for value i
func NewNullBool(i bool) NullBool {
	z := NullBool{}
	z.Valid = true
	z.Bool = i
	return z
}
