package remediation

import (
	"strconv"
	"strings"
	"time"
)

// Suggest returns remediation steps based on context values such as risk and time.
func Suggest(ctx map[string]string) []string {
	var actions []string

	if riskTooHigh(ctx) {
		actions = append(actions, "Require MFA step-up")
	}
	if outsideBusinessHours(ctx) {
		actions = append(actions, "Try again during working hours")
	}
	return actions
}

func riskTooHigh(ctx map[string]string) bool {
	r := ""
	if v, ok := ctx["risk"]; ok {
		r = v
	}
	if v, ok := ctx["risk_score"]; ok {
		r = v
	}
	r = strings.ToLower(r)
	if r == "" {
		return false
	}
	order := map[string]int{"low": 1, "medium": 2, "high": 3}
	if v, ok := order[r]; ok {
		return v > order["medium"]
	}
	if f, err := strconv.ParseFloat(r, 64); err == nil {
		return f > 50
	}
	return false
}

func outsideBusinessHours(ctx map[string]string) bool {
	tStr, ok := ctx["time"]
	if !ok || tStr == "" {
		return false
	}
	t, err := time.Parse("15:04", tStr)
	if err != nil {
		return false
	}
	h := t.Hour()
	return h < 9 || h >= 17
}
