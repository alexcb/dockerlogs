package dockerlogs

import (
	"acb/logparsers/keyvalue"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aybabtme/rgbterm"
)

type LogLevel int

const (
	// Special tokens
	UNKNOWN = iota
	CRITICAL
	ERROR
	WARNING
	INFO
	DEBUG
)

type KeyValue struct {
	Key   string
	Value string
}

type Log struct {
	Level   LogLevel
	Msg     string
	Context []KeyValue
}

func getLevelFromString(s string) LogLevel {
	s = strings.ToLower(s)
	switch s {
	case "CRIT", "CRITICAL", "PANIC":
		return CRITICAL
	case "ERR", "ERROR":
		return ERROR
	case "warn", "warning":
		return WARNING
	case "info":
		return INFO
	case "debug":
		return DEBUG
	default:
		return UNKNOWN
	}
}

func LogLevelToColorString(l LogLevel) string {
	switch l {
	case CRITICAL:
		return rgbterm.BgString("CRT", 255, 0, 0)
	case ERROR:
		return rgbterm.FgString("ERR", 255, 0, 0)
	case WARNING:
		return rgbterm.FgString("WRN", 255, 245, 32)
	case INFO:
		return rgbterm.FgString("INF", 20, 172, 190)
	case DEBUG:
		return rgbterm.FgString("DBG", 221, 28, 119)
	case UNKNOWN:
		return rgbterm.FgString("UNK", 221, 28, 119)
	default:
		panic(fmt.Sprintf("unhandled: %v", l))
	}
}

func parseJsonLog(l string) *Log {
	parsedLog := map[string]interface{}{}
	err := json.Unmarshal([]byte(l), &parsedLog)
	if err != nil {
		return nil
	}
	keyvalues := []KeyValue{}
	msg := ""
	level := LogLevel(UNKNOWN)
	for k, v := range parsedLog {
		switch k {
		case "msg", "message":
			msg = v.(string)
		case "level":
			level = getLevelFromString(v.(string))
		case "time":
			continue
		default:
			keyvalues = append(keyvalues, KeyValue{k, fmt.Sprintf("%v", v)})
		}
	}
	return &Log{
		Level:   level,
		Msg:     msg,
		Context: keyvalues,
	}
}

func parseKeyValueLog(l string) *Log {
	parsedLog, err := keyvalue.NewParser(strings.NewReader(l)).Parse()
	if err != nil {
		return nil
	}
	keyValues := []KeyValue{}
	msg := ""
	level := LogLevel(UNKNOWN)
	for _, kv := range parsedLog {
		switch kv.Key {
		case "msg", "message":
			msg = kv.Value
		case "level":
			level = getLevelFromString(kv.Value)
		case "time":
			continue
		default:
			keyValues = append(keyValues, KeyValue{kv.Key, kv.Value})
		}
	}
	return &Log{
		Level:   level,
		Msg:     msg,
		Context: keyValues,
	}
}

func ParseLog(l string) *Log {

	log := parseJsonLog(l)
	if log != nil {
		return log
	}

	log = parseKeyValueLog(l)
	if log != nil {
		return log
	}

	keyValues := []KeyValue{}
	return &Log{
		Level:   UNKNOWN,
		Msg:     l,
		Context: keyValues,
	}
}

func (l *Log) Format() string {

	buf := []string{}
	buf = append(buf, LogLevelToColorString(l.Level))
	if l.Msg != "" {
		buf = append(buf, rgbterm.FgString(l.Msg, 255, 255, 255))
	}
	for _, x := range l.Context {
		buf = append(buf, rgbterm.FgString(x.Key, 0, 100, 90)+rgbterm.FgString("=", 190, 190, 190)+rgbterm.FgString(x.Value, 120, 120, 120))
	}
	return strings.Join(buf, " ")
}
