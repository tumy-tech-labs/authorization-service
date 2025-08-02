package policy

import "time"

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
