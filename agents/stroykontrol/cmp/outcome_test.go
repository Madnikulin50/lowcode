package cmp

import "testing"

func TestCheckOutcome(t *testing.T) {
	cases := []struct{ verdict, ai, want string }{
		{"", "", OutcomeCheck},
		{VerdictNoMatch, AIUnclear, OutcomeCheck},
		{VerdictNoDocument, "", OutcomeCheck},
		{VerdictResolved, "", OutcomeFixed},
		{VerdictStillDiffers, "", OutcomeNotFixed},
		{"", AIPartial, OutcomePartial},
		{VerdictNoMatch, AIFixed, OutcomeFixed},
		{VerdictResolved, AIFixed, OutcomeFixed},
		{VerdictStillDiffers, AINotFixed, OutcomeNotFixed},
		{VerdictResolved, AINotFixed, OutcomeCheck},  // sources disagree
		{VerdictResolved, AIPartial, OutcomeCheck},   // sources disagree
		{VerdictStillDiffers, AIFixed, OutcomeCheck}, // sources disagree
	}
	for _, c := range cases {
		if got := CheckOutcome(c.verdict, c.ai); got != c.want {
			t.Errorf("CheckOutcome(%q,%q) = %q, want %q", c.verdict, c.ai, got, c.want)
		}
	}
}

func TestResolveDecision(t *testing.T) {
	if k, err := ResolveDecision(OutcomeFixed, OutcomeFixed, ""); err != nil || k != DecisionConfirmed {
		t.Errorf("confirm: %q %v", k, err)
	}
	if _, err := ResolveDecision(OutcomeFixed, OutcomeNotFixed, ""); err == nil {
		t.Error("a changed outcome without a comment must fail")
	}
	if k, err := ResolveDecision(OutcomeFixed, OutcomeNotFixed, "на листе 5 осталось"); err != nil || k != DecisionChanged {
		t.Errorf("change: %q %v", k, err)
	}
	if k, err := ResolveDecision(OutcomeFixed, "", ""); err != nil || k != "" {
		t.Errorf("clear: %q %v", k, err)
	}
	if _, err := ResolveDecision(OutcomeFixed, "bogus", "x"); err == nil {
		t.Error("an unknown outcome must fail")
	}
}
