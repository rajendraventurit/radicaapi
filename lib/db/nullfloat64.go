package db

import (
	"database/sql"
	"encoding/json"
)

// NullFloat64 wraps a sql.NullFloat64 to allow proper json conversion
type NullFloat64 struct {
	sql.NullFloat64
}

// MarshalJSON will convert NullFloat64 to json value or null
func (v NullFloat64) MarshalJSON() ([]byte, error) {
	if v.Valid {
		return json.Marshal(v.Float64)
	}
	return json.Marshal(nil)
}

// UnmarshalJSON will return json encoded for value
func (v *NullFloat64) UnmarshalJSON(data []byte) error {
	// Unmarshalling into a pointer will let us detect null
	var x *float64
	if err := json.Unmarshal(data, &x); err != nil {
		return err
	}
	if x != nil {
		v.Valid = true
		v.Float64 = *x
	} else {
		v.Valid = false
	}
	return nil
}

// NewNullFloat64 returns a valid NullFloat64 for value i
func NewNullFloat64(i float64) NullFloat64 {
	z := NullFloat64{}
	z.Valid = true
	z.Float64 = i
	return z
}
