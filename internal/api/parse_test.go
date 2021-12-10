package api

import (
	"testing"
)

func TestParseDurationInterval(t *testing.T) {
	if min, max, err := parseDuration("12,34"); err != nil {
		t.Fatalf("error: %v", err)
	} else if min != 12 {
		t.Fatalf("invalid minimum duration: %v", min)
	} else if max != 34 {
		t.Fatalf("invalid maximum duration: %v", max)
	}
}

func TestParseDurationIntervalError(t *testing.T) {
	tests := []struct {
		name  string
		value string
	}{
		{
			name:  "empty",
			value: "",
		},
		{
			name:  "one-value",
			value: "12",
		},
		{
			name:  "three-values",
			value: "12,34,56",
		},
		{
			name:  "invalid-min",
			value: "boom,34",
		},
		{
			name:  "invalid-max",
			value: "12,boom",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, _, err := parseDuration(test.value); err == nil {
				t.Fatalf("no error returned")
			}
		})
	}
}
