package types

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"time"
)

// NullString represents a nullable string
type NullString struct {
	sql.NullString
}

// NewNullString creates a new nullable string
func NewNullString(s string) NullString {
	return NullString{sql.NullString{String: s, Valid: true}}
}

// NullStringFrom creates a valid nullable string
func NullStringFrom(s string) NullString {
	return NullString{sql.NullString{String: s, Valid: true}}
}

// NullStringFromPtr creates a nullable string from pointer
func NullStringFromPtr(s *string) NullString {
	if s == nil {
		return NullString{sql.NullString{Valid: false}}
	}
	return NullString{sql.NullString{String: *s, Valid: true}}
}

// FromNullString converts sql.NullString to types.NullString
// Deprecated: Use types.NullString directly in adapter row structs instead of sql.NullString.
// This function remains for backward compatibility only.
func FromNullString(ns sql.NullString) NullString {
	return NullString{ns}
}

// ToNullString converts types.NullString to sql.NullString
// Deprecated: Use types.NullString directly in adapter row structs instead of sql.NullString.
// This function remains for backward compatibility only.
func ToNullString(s NullString) sql.NullString {
	return s.NullString
}

// MarshalJSON implements json.Marshaler
func (s NullString) MarshalJSON() ([]byte, error) {
	if !s.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(s.String)
}

// UnmarshalJSON implements json.Unmarshaler
func (s *NullString) UnmarshalJSON(data []byte) error {
	var v *string
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	if v == nil {
		s.Valid = false
		return nil
	}
	s.String = *v
	s.Valid = true
	return nil
}

// NullTime represents a nullable time
type NullTime struct {
	sql.NullTime
}

// NewNullTime creates a new nullable time
func NewNullTime(t time.Time) NullTime {
	return NullTime{sql.NullTime{Time: t, Valid: true}}
}

// NullTimeFrom creates a valid nullable time
func NullTimeFrom(t time.Time) NullTime {
	return NullTime{sql.NullTime{Time: t, Valid: true}}
}

// NullTimeFromPtr creates a nullable time from pointer
func NullTimeFromPtr(t *time.Time) NullTime {
	if t == nil {
		return NullTime{sql.NullTime{Valid: false}}
	}
	return NullTime{sql.NullTime{Time: *t, Valid: true}}
}

// FromNullTime converts sql.NullTime to types.NullTime
// Deprecated: Use types.NullTime directly in adapter row structs instead of sql.NullTime.
// This function remains for backward compatibility only.
func FromNullTime(nt sql.NullTime) NullTime {
	return NullTime{nt}
}

// ToNullTime converts types.NullTime to sql.NullTime
// Deprecated: Use types.NullTime directly in adapter row structs instead of sql.NullTime.
// This function remains for backward compatibility only.
func ToNullTime(t NullTime) sql.NullTime {
	return t.NullTime
}

// MarshalJSON implements json.Marshaler
func (t NullTime) MarshalJSON() ([]byte, error) {
	if !t.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(t.Time)
}

// UnmarshalJSON implements json.Unmarshaler
func (t *NullTime) UnmarshalJSON(data []byte) error {
	var v *time.Time
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	if v == nil {
		t.Valid = false
		return nil
	}
	t.Time = *v
	t.Valid = true
	return nil
}

// NullInt64 represents a nullable int64
type NullInt64 struct {
	sql.NullInt64
}

// NewNullInt64 creates a new nullable int64
func NewNullInt64(i int64, valid bool) NullInt64 {
	return NullInt64{sql.NullInt64{Int64: i, Valid: valid}}
}

// NullInt64From creates a valid nullable int64
func NullInt64From(i int64) NullInt64 {
	return NullInt64{sql.NullInt64{Int64: i, Valid: true}}
}

// NullInt64FromPtr creates a nullable int64 from pointer
func NullInt64FromPtr(i *int64) NullInt64 {
	if i == nil {
		return NullInt64{sql.NullInt64{Valid: false}}
	}
	return NullInt64{sql.NullInt64{Int64: *i, Valid: true}}
}

// MarshalJSON implements json.Marshaler
func (i NullInt64) MarshalJSON() ([]byte, error) {
	if !i.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(i.Int64)
}

// UnmarshalJSON implements json.Unmarshaler
func (i *NullInt64) UnmarshalJSON(data []byte) error {
	var v *int64
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	if v == nil {
		i.Valid = false
		return nil
	}
	i.Int64 = *v
	i.Valid = true
	return nil
}

// NullBool represents a nullable bool
type NullBool struct {
	sql.NullBool
}

// NewNullBool creates a new nullable bool
func NewNullBool(b bool, valid bool) NullBool {
	return NullBool{sql.NullBool{Bool: b, Valid: valid}}
}

// NullBoolFrom creates a valid nullable bool
func NullBoolFrom(b bool) NullBool {
	return NullBool{sql.NullBool{Bool: b, Valid: true}}
}

// NullBoolFromPtr creates a nullable bool from pointer
func NullBoolFromPtr(b *bool) NullBool {
	if b == nil {
		return NullBool{sql.NullBool{Valid: false}}
	}
	return NullBool{sql.NullBool{Bool: *b, Valid: true}}
}

// MarshalJSON implements json.Marshaler
func (b NullBool) MarshalJSON() ([]byte, error) {
	if !b.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(b.Bool)
}

// UnmarshalJSON implements json.Unmarshaler
func (b *NullBool) UnmarshalJSON(data []byte) error {
	var v *bool
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	if v == nil {
		b.Valid = false
		return nil
	}
	b.Bool = *v
	b.Valid = true
	return nil
}

// NullFloat64 represents a nullable float64
type NullFloat64 struct {
	sql.NullFloat64
}

// NewNullFloat64 creates a new nullable float64
func NewNullFloat64(f float64, valid bool) NullFloat64 {
	return NullFloat64{sql.NullFloat64{Float64: f, Valid: valid}}
}

// NullFloat64From creates a valid nullable float64
func NullFloat64From(f float64) NullFloat64 {
	return NullFloat64{sql.NullFloat64{Float64: f, Valid: true}}
}

// NullFloat64FromPtr creates a nullable float64 from pointer
func NullFloat64FromPtr(f *float64) NullFloat64 {
	if f == nil {
		return NullFloat64{sql.NullFloat64{Valid: false}}
	}
	return NullFloat64{sql.NullFloat64{Float64: *f, Valid: true}}
}

// MarshalJSON implements json.Marshaler
func (f NullFloat64) MarshalJSON() ([]byte, error) {
	if !f.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(f.Float64)
}

// UnmarshalJSON implements json.Unmarshaler
func (f *NullFloat64) UnmarshalJSON(data []byte) error {
	var v *float64
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	if v == nil {
		f.Valid = false
		return nil
	}
	f.Float64 = *v
	f.Valid = true
	return nil
}

// Ensure types implement sql.Scanner and driver.Valuer
var (
	_ sql.Scanner   = (*NullString)(nil)
	_ driver.Valuer = (*NullString)(nil)
	_ sql.Scanner   = (*NullTime)(nil)
	_ driver.Valuer = (*NullTime)(nil)
	_ sql.Scanner   = (*NullInt64)(nil)
	_ driver.Valuer = (*NullInt64)(nil)
	_ sql.Scanner   = (*NullBool)(nil)
	_ driver.Valuer = (*NullBool)(nil)
	_ sql.Scanner   = (*NullFloat64)(nil)
	_ driver.Valuer = (*NullFloat64)(nil)
)
