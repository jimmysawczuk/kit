package timestamp_test

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/jimmysawczuk/kit/timestamp"
	"github.com/stretchr/testify/require"
)

func TestTimestamp(t *testing.T) {
	tests := []struct {
		name  string
		t     func() *time.Time
		valid bool
		json  string
	}{
		{
			name: "VALID",
			t: func() *time.Time {
				t := new(time.Time)
				*t = time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC)
				return t
			},
			valid: true,
			json:  `"2006-01-02T15:04:05Z"`,
		},
		{
			name: "NIL",
			t:    func() *time.Time { return nil },
			json: `null`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var err error

			ts := timestamp.NewFromPtr(test.t())

			require.Equal(t, test.valid, ts.Valid)

			jsts, _ := json.Marshal(ts)
			require.Equal(t, test.json, string(jsts))

			uts := timestamp.Timestamp{}

			err = json.Unmarshal(jsts, &uts)
			require.NoError(t, err)

			require.Equal(t, ts, uts)
		})
	}
}

func TestScan(t *testing.T) {
	tests := []struct {
		in     interface{}
		out    timestamp.Timestamp
		hasErr bool
	}{
		{
			in:  nil,
			out: timestamp.Timestamp{},
		},
		{
			in:     []byte(nil),
			out:    timestamp.Timestamp{},
			hasErr: true,
		},
		{
			in:  []byte(`0000-00-00 00:00:00`),
			out: timestamp.Timestamp{},
		},
		{
			in: []byte(`2006-01-02 15:04:05`),
			out: timestamp.Timestamp{
				Time:  time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC),
				Valid: true,
			},
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("TEST_%d", i+1), func(t *testing.T) {
			ts := timestamp.Timestamp{}
			err := ts.Scan(test.in)
			require.Equal(t, test.hasErr, err != nil)

			require.Equal(t, test.out, ts)
		})
	}
}

func TestValue(t *testing.T) {
	tests := []struct {
		in  timestamp.Timestamp
		out driver.Value
	}{
		{
			in:  timestamp.Timestamp{},
			out: driver.Value(nil),
		},
		{
			in:  timestamp.New(time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC)),
			out: driver.Value("2006-01-02 15:04:05"),
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("TEST_%d", i+1), func(t *testing.T) {
			v, err := test.in.Value()
			require.NoError(t, err)
			require.Equal(t, test.out, v)
		})
	}
}
