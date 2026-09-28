package updater

import (
	"testing"
)

func TestCompareSemVer(t *testing.T) {
	tests := []struct {
		name     string
		v1       string
		v2       string
		expected int
	}{
		{
			name:     "patch difference",
			v1:       "1.0.1",
			v2:       "1.0.0",
			expected: 1,
		},
		{
			name:     "patch difference inverted",
			v1:       "1.0.0",
			v2:       "1.0.1",
			expected: -1,
		},
		{
			name:     "with v prefix and minor difference",
			v1:       "v1.2.0",
			v2:       "1.1.9",
			expected: 1,
		},
		{
			name:     "equal versions",
			v1:       "1.0.0",
			v2:       "1.0.0",
			expected: 0,
		},
		{
			name:     "equal versions with prefixes and spaces",
			v1:       " v1.0.0 ",
			v2:       "1.0.0",
			expected: 0,
		},
		{
			name:     "two-digit minor vs single-digit minor",
			v1:       "1.10.0",
			v2:       "1.9.0",
			expected: 1,
		},
		{
			name:     "two-digit minor vs single-digit minor inverted",
			v1:       "1.9.0",
			v2:       "1.10.0",
			expected: -1,
		},
		{
			name:     "major difference",
			v1:       "2.0.0",
			v2:       "1.99.99",
			expected: 1,
		},
		{
			name:     "prerelease vs stable",
			v1:       "1.0.0",
			v2:       "1.0.0-rc1",
			expected: 1,
		},
		{
			name:     "prerelease vs stable inverted",
			v1:       "1.0.0-rc1",
			v2:       "1.0.0",
			expected: -1,
		},
		{
			name:     "prerelease vs prerelease",
			v1:       "1.0.0-rc2",
			v2:       "1.0.0-rc1",
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := CompareSemVer(tt.v1, tt.v2)
			if res != tt.expected {
				t.Errorf("CompareSemVer(%q, %q) = %d; expected %d", tt.v1, tt.v2, res, tt.expected)
			}
		})
	}
}

func TestMatchPlatformAsset(t *testing.T) {
	assets := []ReleaseAsset{
		{
			Name:        "AetherGrok-1.0.1-macOS-arm64.dmg",
			DownloadURL: "https://example.com/AetherGrok-1.0.1-macOS-arm64.dmg",
			Size:        1024000,
		},
		{
			Name:        "AetherGrok-1.0.1-macOS-x64.dmg",
			DownloadURL: "https://example.com/AetherGrok-1.0.1-macOS-x64.dmg",
			Size:        1024000,
		},
		{
			Name:        "AetherGrok-1.0.1-windows-x64-setup.exe",
			DownloadURL: "https://example.com/AetherGrok-1.0.1-windows-x64-setup.exe",
			Size:        1024000,
		},
		{
			Name:        "AetherGrok-1.0.1-windows-arm64-setup.exe",
			DownloadURL: "https://example.com/AetherGrok-1.0.1-windows-arm64-setup.exe",
			Size:        1024000,
		},
	}

	tests := []struct {
		name         string
		osName       string
		arch         string
		expectedName string
	}{
		{
			name:         "macOS arm64",
			osName:       "darwin",
			arch:         "arm64",
			expectedName: "AetherGrok-1.0.1-macOS-arm64.dmg",
		},
		{
			name:         "macOS x64",
			osName:       "darwin",
			arch:         "amd64",
			expectedName: "AetherGrok-1.0.1-macOS-x64.dmg",
		},
		{
			name:         "Windows x64",
			osName:       "windows",
			arch:         "amd64",
			expectedName: "AetherGrok-1.0.1-windows-x64-setup.exe",
		},
		{
			name:         "Windows arm64",
			osName:       "windows",
			arch:         "arm64",
			expectedName: "AetherGrok-1.0.1-windows-arm64-setup.exe",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matched := MatchPlatformAsset(assets, tt.osName, tt.arch)
			if matched == nil {
				t.Fatalf("expected asset %q, got nil", tt.expectedName)
			}
			if matched.Name != tt.expectedName {
				t.Errorf("expected asset name %q, got %q", tt.expectedName, matched.Name)
			}
		})
	}
}

func TestExtractHighlightsFromMarkdown(t *testing.T) {
	md := `
# Release Notes v1.0.1
Here are some highlights:
- Pure-Go persistent JSONL storage engine
* Full support for macOS Apple Silicon and Windows
• Integrated terminal dock with split tabs

Other notes that are not bullet points.
`
	highlights := ExtractHighlightsFromMarkdown(md)
	if len(highlights) != 3 {
		t.Fatalf("expected 3 highlights, got %d: %v", len(highlights), highlights)
	}

	expected := []string{
		"Pure-Go persistent JSONL storage engine",
		"Full support for macOS Apple Silicon and Windows",
		"Integrated terminal dock with split tabs",
	}

	for i, exp := range expected {
		if highlights[i] != exp {
			t.Errorf("highlight[%d] = %q; expected %q", i, highlights[i], exp)
		}
	}
}
