package omega

import (
	"encoding/json"
	"time"
)

// LogLevel defines the types of log level available on the api
type LogLevel string

// Logs levels available
const (
	LogLevelDebug LogLevel = "DEBUG"
	LogLevelInfo  LogLevel = "INFO"
	LogLevelWarn  LogLevel = "WARNING"
	LogLevelError LogLevel = "ERROR"
	LogLevelFatal LogLevel = "FATAL"
)

// LogEntry represents an activity record
type LogEntry struct {
	MrID         string                 `json:"mrId"`
	SpanID       string                 `json:"spanId,omitempty"`
	TraceID      string                 `json:"traceId,omitempty"`
	CreationDate time.Time              `json:"-"`
	RecordDate   time.Time              `json:"-"`
	Level        LogLevel               `json:"level,omitempty"`
	Message      string                 `json:"message,omitempty"`
	Properties   map[string]interface{} `json:"properties,omitempty"`
}

type logEntryInternal LogEntry // just to avoid a marshall loop

type jsonLogEntry struct {
	logEntryInternal
	CreationDate int64 `json:"creationDate"`
	RecordDate   int64 `json:"recordDate"`
}

func (m *LogEntry) MarshalJSON() ([]byte, error) {
	return json.Marshal(
		jsonLogEntry{
			logEntryInternal: logEntryInternal(*m),
			CreationDate:     m.CreationDate.UnixNano(),
			RecordDate:       m.RecordDate.UnixNano(),
		},
	)
}

func (m *LogEntry) UnmarshalJSON(data []byte) error {
	var v jsonLogEntry
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	*m = LogEntry(v.logEntryInternal)
	m.CreationDate = time.Unix(0, v.CreationDate)
	m.RecordDate = time.Unix(0, v.RecordDate)
	return nil
}
