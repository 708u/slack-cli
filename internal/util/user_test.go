package util

import "testing"

func TestIsUserID(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"U0123456789", true},
		{"UABCDEF012", true},
		{"W0123456789", true},
		{"C0123456789", false},
		{"D0123456789", false},
		{"G0123456789", false},
		{"U123", false},
		{"bob", false},
		{"@alice", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := IsUserID(tt.input); got != tt.want {
				t.Errorf("IsUserID(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
