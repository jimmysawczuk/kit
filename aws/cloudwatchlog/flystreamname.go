package cloudwatchlog

import (
	"os"

	"github.com/oklog/ulid/v2"
)

// FlyStreamName returns a CloudWatch log stream name suitable for a process
// running on Fly.io: the Fly machine ID, which is unique and stable per
// running machine. Outside Fly (local dev, CI) FLY_MACHINE_ID is unset, so
// it falls back to hostname+ulid -- otherwise two processes on the same host
// would write to a stream named after the shared hostname and stomp on each
// other's sequence tokens.
func FlyStreamName() string {
	if id := os.Getenv("FLY_MACHINE_ID"); id != "" {
		return id
	}

	host, _ := os.Hostname()
	if host == "" {
		host = "unknown"
	}

	return host + "-" + ulid.Make().String()
}
