package dockerlogs

import (
	"acb/logparsers/keyvalue"
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
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

type KeyValues []KeyValue

func (s KeyValues) Len() int {
	return len(s)
}
func (s KeyValues) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s KeyValues) Less(i, j int) bool {
	return s[i].Key < s[j].Key
}

type Log struct {
	Level   LogLevel
	Msg     string
	Context KeyValues
}

func getLevelFromString(s string) LogLevel {
	s = strings.ToLower(s)
	switch s {
	case "crit", "critical", "panic":
		return CRITICAL
	case "err", "error":
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

func formatValue(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case int:
		return strconv.Itoa(v)
	case int64:
		return strconv.FormatInt(v, 10)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case []interface{}:
		var buffer bytes.Buffer
		buffer.WriteString("[")
		for i, x := range v {
			if i != 0 {
				buffer.WriteString(" ")
			}
			buffer.WriteString(formatValue(x))
		}
		buffer.WriteString("]")
		return buffer.String()
	case map[interface{}]interface{}:
		var buffer bytes.Buffer
		buffer.WriteString("{")
		first := true
		for k, val := range v {
			if first {
				first = false
			} else {
				buffer.WriteString(" ")
			}
			buffer.WriteString(formatValue(k))
			buffer.WriteString(":")
			buffer.WriteString(formatValue(val))
		}
		buffer.WriteString("}")
		return buffer.String()
	default:
		return fmt.Sprintf("%v", v)
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
			keyvalues = append(keyvalues, KeyValue{k, formatValue(v)})
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
	sort.Sort(l.Context)
	for _, x := range l.Context {
		buf = append(buf, rgbterm.FgString(x.Key, 0, 100, 90)+rgbterm.FgString("=", 190, 190, 190)+rgbterm.FgString(x.Value, 120, 120, 120))
	}
	return strings.Join(buf, " ")
}
