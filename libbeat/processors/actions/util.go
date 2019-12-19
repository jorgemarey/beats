package actions

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/elastic/beats/libbeat/beat"
	"globaldevtools.bbva.com/entsec/semaas.git/client/omega"
)

var re = regexp.MustCompile(`^(?i:[A-Za-z0-9\_]|[A-Za-z0-9][A-Za-z0-9\-\_]{0,61}[a-z0-9])$`)

func parseItem(parent, key string, item interface{}, event *beat.Event) {
	// Replace dots with underscores
	key = strings.ReplaceAll(key, ".", "_")
	// Validate key without special charactres
	if !re.MatchString(key) {
		return
	}
	switch concreteVal := item.(type) {
	case map[string]interface{}:
		parseMap(parent, key, item.(map[string]interface{}), event)
	case []interface{}:
		parseArray(parent, key, item.([]interface{}), event)
	case string:
		if concreteVal != "" {
			event.PutValue(parent+"."+key, concreteVal)
		}
	default:
		if concreteVal != nil {
			event.PutValue(parent+"."+key, concreteVal)
		}
	}
}

func parseMap(parent, prevKey string, aMap map[string]interface{}, event *beat.Event) {
	for key, val := range aMap {
		// Validate key without special charactres
		if !re.MatchString(key) {
			continue
		}
		mapKey := key
		if prevKey != "" {
			mapKey = fmt.Sprintf("%s_%s", prevKey, key)
		}
		parseItem(parent, mapKey, val, event)
	}
}

func parseArray(parent, prevKey string, anArray []interface{}, event *beat.Event) {
	ok := false
	for i, val := range anArray {
		switch val.(type) {
		case map[string]interface{}:
			parseMap(parent, fmt.Sprintf("%s_%d", prevKey, i), val.(map[string]interface{}), event)
		case []interface{}:
			parseArray(parent, fmt.Sprintf("%s_%d", prevKey, i), val.([]interface{}), event)
		default:
			ok = true
			break
		}
	}
	if ok {
		event.PutValue(parent+"."+prevKey, anArray)
	}
}

// this maps a severity -> LogLevel
func getLevelStr(lvl string) omega.LogLevel {
	switch lvl {
	case "notice":
		return omega.LogLevelInfo
	case "err":
		return omega.LogLevelError
	case "warn":
		return omega.LogLevelWarn
	case "emerg", "panic", "crit", "alert":
		return omega.LogLevelFatal
	case "trace":
		return omega.LogLevelDebug
	default: // warning, error, info, debug
		return omega.LogLevel(strings.ToUpper(lvl))
	}
}

// this maps a severity -> LogLevel
func getLevel(lvl int) omega.LogLevel {
	switch lvl {
	case 0, 1, 2:
		return omega.LogLevelFatal
	case 3:
		return omega.LogLevelError
	case 4:
		return omega.LogLevelWarn
	case 5, 6:
		return omega.LogLevelInfo
	case 7:
		return omega.LogLevelDebug
	default:
		return omega.LogLevelInfo
	}
}
