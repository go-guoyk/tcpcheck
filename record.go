package main

import "time"

const (
	ConnectionShort = "short"
	ConnectionLong  = "long"

	ActionConnect     = "connect"
	ActionRoundTrip   = "round-trip"
	ActionTransfer10m = "transfer-10m"
)

type Record struct {
	Timestamp      string `json:"timestamp"`
	Source         string `json:"source"`
	Destination    string `json:"destination"`
	ConnectionType string `json:"connection_type"` // long, short
	ConnectionID   string `json:"connection_id"`   // connection id
	Action         string `json:"action"`          // connect, round-trip, transfer-10m
	Success        bool   `json:"success"`         // action success or not
	Duration       int64  `json:"duration"`        // milliseconds
	Error          string `json:"error"`           // error message, if any
}

func (r Record) CloneSuccess(action string, duration int64) Record {
	nr := r
	nr.Timestamp = time.Now().Format(time.RFC3339Nano)
	nr.Action = action
	nr.Success = true
	nr.Duration = duration
	return nr
}

func (r Record) CloneFailure(action string, duration int64, error string) Record {
	nr := r
	nr.Timestamp = time.Now().Format(time.RFC3339Nano)
	nr.Action = action
	nr.Success = false
	nr.Duration = duration
	nr.Error = error
	return nr
}
