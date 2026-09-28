package grokrunner

import (
	"testing"
)

func TestParseGrokVersion(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"grok 1.0.41 (4220f3b224a6) [stable]", "1.0.41"},
		{"grok v2.0.0", "2.0.0"},
		{"Grok 0.9.15", "0.9.15"},
		{"custom build 1.2.3", "custom build 1.2.3"},
	}

	for _, tc := range tests {
		got := parseGrokVersion(tc.input)
		if got != tc.expected {
			t.Errorf("parseGrokVersion(%q) = %q; want %q", tc.input, got, tc.expected)
		}
	}
}

func TestDetectGrokInstallation(t *testing.T) {
	status := DetectGrokInstallation()
	t.Logf("Detected status: %+v", status)
	if status.Platform == "" {
		t.Errorf("expected platform to be populated, got empty")
	}
}
