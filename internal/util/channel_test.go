package util

import "testing"

func TestIsChannelID(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"C0123456789", true},
		{"D0123456789", true},
		{"G0123456789", true},
		{"C01234567", true},  // 8 chars after prefix matches {8,}
		{"C0123456", false},  // only 7 chars after prefix
		{"C012345678", true}, // 9 chars after prefix
		{"general", false},
		{"c0123456789", false}, // lowercase
		{"X0123456789", false}, // wrong prefix
		{"", false},
	}
	for _, tt := range tests {
		got := IsChannelID(tt.input)
		if got != tt.expected {
			t.Errorf("IsChannelID(%q) = %v, want %v", tt.input, got, tt.expected)
		}
	}
}

func TestFormatChannelName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", "#unknown"},
		{"general", "#general"},
		{"#general", "#general"},
	}
	for _, tt := range tests {
		got := FormatChannelName(tt.input)
		if got != tt.expected {
			t.Errorf("FormatChannelName(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestGetChannelTypes(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"public", "public_channel"},
		{"private", "private_channel"},
		{"im", "im"},
		{"mpim", "mpim"},
		{"all", "public_channel,private_channel,mpim,im"},
		{"unknown", "public_channel"},
	}
	for _, tt := range tests {
		got := GetChannelTypes(tt.input)
		if got != tt.expected {
			t.Errorf("GetChannelTypes(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}
