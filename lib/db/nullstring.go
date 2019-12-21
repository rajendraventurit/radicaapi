package db

import (
	"database/sql"
	"encoding/json"
)

// NullString wraps a sql.NullString to allow proper json conversion
type NullString struct {
	sql.NullString
}

// MarshalJSON will convert NullString to json value or null
func (v NullString) MarshalJSON() ([]byte, error) {
	if v.Valid {
		return json.Marshal(v.String)
	}
	return json.Marshal(nil)
}

// UnmarshalJSON will return json encoded for value
func (v *NullString) UnmarshalJSON(data []byte) error {
	// Unmarshalling into a pointer will let us detect null
	var x *string
	if err := json.Unmarshal(data, &x); err != nil {
		return err
	}
	if x != nil {
		v.Valid = true
		v.String = *x
	} else {
		v.Valid = false
	}
	return nil
}

// NewNullString returns a valid NullString for value i
func NewNullString(s string) NullString {
	z := NullString{}
	z.Valid = true
	z.String = s
	return z
}

// AsNullString converts a string into a sql.NullString
func AsNullString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: true}
}
