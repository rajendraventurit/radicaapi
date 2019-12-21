package db

import (
	"database/sql"
	"encoding/json"
)

// NullInt64 wraps a sql.NullInt64 to allow proper json conversion
type NullInt64 struct {
	sql.NullInt64
}

// MarshalJSON will convert NullInt64 to json value or null
func (v NullInt64) MarshalJSON() ([]byte, error) {
	if v.Valid {
		return json.Marshal(v.Int64)
	}
	return json.Marshal(nil)
}

// UnmarshalJSON will return json encoded for value
func (v *NullInt64) UnmarshalJSON(data []byte) error {
	// Unmarshalling into a pointer will let us detect null
	var x *int64
	if err := json.Unmarshal(data, &x); err != nil {
		return err
	}
	if x != nil {
		v.Valid = true
		v.Int64 = *x
	} else {
		v.Valid = false
	}
	return nil
}

// NewNullInt64 returns a valid NullInt64 for value i
func NewNullInt64(i int64) NullInt64 {
	z := NullInt64{}
	z.Valid = true
	z.Int64 = i
	return z
}

// AsNullInt64 converts a int64 into a sql.NullInt64
func AsNullInt64(i int64) sql.NullInt64 {
	return sql.NullInt64{Int64: i, Valid: true}
}
