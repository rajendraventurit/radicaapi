package db

import (
	"encoding/json"
	"time"

	"github.com/go-sql-driver/mysql"
)

// NullTime wraps a sql.NullTime to allow proper json conversion
type NullTime struct {
	mysql.NullTime
}

// MarshalJSON will convert NullTime to json value or null
func (v NullTime) MarshalJSON() ([]byte, error) {
	if v.Valid {
		return json.Marshal(v.Time)
	}
	return json.Marshal(nil)
}

// UnmarshalJSON will return json encoded for value
func (v *NullTime) UnmarshalJSON(data []byte) error {
	// Unmarshalling into a pointer will let us detect null
	var x *time.Time
	if err := json.Unmarshal(data, &x); err != nil {
		return err
	}
	if x != nil {
		v.Valid = true
		v.Time = *x
	} else {
		v.Valid = false
	}
	return nil
}

// NewNullTime returns a valid NullTime for value i
func NewNullTime(i time.Time) NullTime {
	z := NullTime{}
	z.Valid = true
	z.Time = i
	return z
}

// AsNullTime converts a time into a sql.NullTime
func AsNullTime(t time.Time) mysql.NullTime {
	return mysql.NullTime{Time: t, Valid: true}
}
