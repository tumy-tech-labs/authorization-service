package remediation

import "testing"

func TestSuggestRisk(t *testing.T) {
	ctx := map[string]string{"risk": "high"}
	res := Suggest(ctx)
	if len(res) != 1 || res[0] != "Require MFA step-up" {
		t.Fatalf("expected MFA remediation, got %v", res)
	}
}

func TestSuggestBusinessHours(t *testing.T) {
	ctx := map[string]string{"time": "20:00"}
	res := Suggest(ctx)
	if len(res) != 1 || res[0] != "Try again during working hours" {
		t.Fatalf("expected working hours remediation, got %v", res)
	}
}

func TestSuggestNone(t *testing.T) {
	ctx := map[string]string{"risk": "low", "time": "10:00"}
	res := Suggest(ctx)
	if len(res) != 0 {
		t.Fatalf("expected no remediation, got %v", res)
	}
}
