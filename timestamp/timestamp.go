package timestamp

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// Timestamp wraps a time.Time in a struct with a Valid bool so we can safely handle NULL values.
type Timestamp struct {
	Time  time.Time
	Valid bool
}

// New returns a Timestamp with the provided time.Time and Valid set to true.
func New(t time.Time) Timestamp {
	return Timestamp{
		Time:  t,
		Valid: true,
	}
}

// NewFromPtr wraps New to ensure that the provided *time.Time is not nil before deferencing it and creating
// a new Timestamp.
func NewFromPtr(t *time.Time) Timestamp {
	if t == nil {
		return Timestamp{}
	}

	return New(*t)
}

// Now wraps time.Now with a Timestamp.
func Now() Timestamp {
	return New(time.Now().UTC())
}

// Parse wraps time.Parse. It'll return an error if the underlying call to time.Parse returns an error.
func Parse(layout string, val string) (Timestamp, error) {
	t, err := time.Parse(layout, val)
	if err != nil {
		return Timestamp{}, fmt.Errorf("parse: %w", err)
	}

	return New(t), nil
}

// ParseInLocation wraps time.ParseInLocation. It'll return an error if the underlying call to time.ParseInLocation
// returns an error.
func ParseInLocation(layout string, val string, loc *time.Location) (Timestamp, error) {
	t, err := time.ParseInLocation(layout, val, loc)
	if err != nil {
		return Timestamp{}, fmt.Errorf("parse in location: %w", err)
	}

	return New(t), nil
}

// Must panics if err != nil.
func Must(t Timestamp, err error) Timestamp {
	if err != nil {
		panic(err)
	}

	return t
}

// MarshalJSON implements json.Marshaler. If the Timestamp is valid, it redirects to the time.Time's JSON marshaler;
// if it is not valid, it returns json.Marshal(nil).
func (t Timestamp) MarshalJSON() ([]byte, error) {
	if !t.Valid {
		return json.Marshal(nil)
	}

	return json.Marshal(t.Time.UTC().Truncate(1 * time.Second))
}

// UnmarshalJSON implements json.Unmarshaler. If the incoming []byte == "null", it'll return a Timestamp with Valid
// set to false. Otherwise, it'll attempt to parse the time.Time and return a Timestamp with Valid set to true if
// it's able.
func (t *Timestamp) UnmarshalJSON(in []byte) error {
	if string(in) == "null" {
		*t = Timestamp{}
		return nil
	}

	target := time.Time{}

	err := json.Unmarshal(in, &target)
	if err != nil {
		return err
	}

	t.Time = target
	t.Valid = true

	return nil
}

// IsNull returns true if Timestamp.Valid == false.
func (t *Timestamp) IsNull() bool {
	return !t.Valid
}

// Value implements driver.Valuer.
func (t Timestamp) Value() (driver.Value, error) {
	if t.IsNull() {
		return nil, nil
	}

	return t.Time.In(time.UTC).Format("2006-01-02 15:04:05"), nil
}

// Scan implements driver.Scanner.
func (t *Timestamp) Scan(in interface{}) error {
	if in == nil {
		t.Time = time.Time{}
		t.Valid = false
		return nil
	}

	switch typed := in.(type) {
	case nil:
		t.Time = time.Time{}
		t.Valid = false
		return nil

	case []byte:
		var err error

		if string(typed) == "0000-00-00 00:00:00" {
			t.Time = time.Time{}
			t.Valid = false
			return nil
		}

		t.Time, err = time.ParseInLocation("2006-01-02 15:04:05", string(typed), time.UTC)
		if err != nil {
			return err
		}

		t.Valid = true
		return nil
	}

	return fmt.Errorf("invalid format: %T", in)
}

// String implements fmt.Stringer.
func (t Timestamp) String() string {
	if t.IsNull() {
		return "<nil>"
	}

	return t.Time.In(time.UTC).Format("2006-01-02 15:04:05")
}
