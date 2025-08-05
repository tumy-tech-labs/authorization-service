package policy

import (
	"strconv"
	"strings"
	"time"
)

// now is a variable for mocking current time in tests.
var now = time.Now

// evaluateConditions checks whether all policy conditions are satisfied using
// the provided environment values.

func evaluateConditions(policyConds map[string]string, env map[string]string) bool {
	if len(policyConds) == 0 {
		return true
	}
	for key, expected := range policyConds {
		var res bool
		switch key {
		case "time":
			res = evaluateTimeCondition(expected, env)
		default:
			if v, ok := env[key]; ok {
				res = v == expected
			} else {
				res = false
			}
		}
		if !res {
			return false
		}
	}
	return true
}

// evaluateWhen evaluates a list of boolean expressions against the environment.
// Supported operators are ==, <, and >. Expressions must reference context
// values using the form `context.key`.
func evaluateWhen(exprs []string, env map[string]string) bool {
	if len(exprs) == 0 {
		return true
	}
	for _, expr := range exprs {
		if !evaluateExpression(expr, env) {
			return false
		}
	}
	return true
}

// evaluateExpression parses and evaluates a single expression.
func evaluateExpression(expr string, env map[string]string) bool {
	expr = strings.TrimSpace(expr)
	var op string
	var parts []string
	if strings.Contains(expr, "==") {
		op = "=="
		parts = strings.SplitN(expr, "==", 2)
	} else if strings.Contains(expr, ">") {
		op = ">"
		parts = strings.SplitN(expr, ">", 2)
	} else if strings.Contains(expr, "<") {
		op = "<"
		parts = strings.SplitN(expr, "<", 2)
	} else {
		return false
	}
	if len(parts) != 2 {
		return false
	}
	left := strings.TrimSpace(parts[0])
	right := strings.TrimSpace(parts[1])
	right = strings.Trim(right, "'\"")
	if !strings.HasPrefix(left, "context.") {
		return false
	}
	key := strings.TrimPrefix(left, "context.")
	val, ok := env[key]
	if !ok {
		return false
	}
	switch op {
	case "==":
		return val == right
	case "<", ">":
		return compareValues(val, right, op)
	}
	return false
}

// compareValues compares two values using the provided operator. It attempts
// numeric comparison, then known risk level ordering, and finally falls back to
// lexical string comparison.
func compareValues(left, right, op string) bool {
	if lf, err := strconv.ParseFloat(left, 64); err == nil {
		if rf, err := strconv.ParseFloat(right, 64); err == nil {
			switch op {
			case "<":
				return lf < rf
			case ">":
				return lf > rf
			}
		}
	}
	order := map[string]int{"low": 1, "medium": 2, "high": 3}
	if lv, ok := order[strings.ToLower(left)]; ok {
		if rv, ok := order[strings.ToLower(right)]; ok {
			switch op {
			case "<":
				return lv < rv
			case ">":
				return lv > rv
			}
		}
	}
	switch op {
	case "<":
		return left < right
	case ">":
		return left > right
	}
	return false
}

// evaluateTimeCondition evaluates the "time" condition. The expected value
// "business-hours" means the time must be between 9:00 and 17:00.
// The current time is taken from env["time"] in HH:MM format, or time.Now() if
// not provided.
func evaluateTimeCondition(expected string, env map[string]string) bool {
	if expected != "business-hours" {
		return false
	}
	var t time.Time
	if env != nil {
		if ts, ok := env["time"]; ok {
			if parsed, err := time.Parse("15:04", ts); err == nil {
				t = parsed
			}
		}
	}
	if t.IsZero() {
		t = now()
	}
	hour := t.Hour()
	return hour >= 9 && hour < 17
}
