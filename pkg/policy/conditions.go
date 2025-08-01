package policy

import "time"

// now is a variable for mocking current time in tests.
var now = time.Now

// evaluateConditions checks whether all policy conditions are satisfied using the
// provided environment values. It returns a map of condition evaluation results
// and a boolean indicating if all conditions passed.
func evaluateConditions(policyConds map[string]string, env map[string]string) (map[string]bool, bool) {
	if len(policyConds) == 0 {
		return nil, true
	}
	results := make(map[string]bool)
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
		results[key] = res
		if !res {
			return results, false
		}
	}
	return results, true
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
