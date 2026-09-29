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

func TestMatchPlatformAsset_WindowsPriorities(t *testing.T) {
	// Scenario 1: Both NSIS installer and portable zip exist for Windows x64 -> Prioritize NSIS installer
	assetsWithSetupAndZip := []ReleaseAsset{
		{Name: "aethergrok-windows-amd64.zip", DownloadURL: "https://example.com/aethergrok-windows-amd64.zip"},
		{Name: "AetherGrok-Setup.exe", DownloadURL: "https://example.com/AetherGrok-Setup.exe"},
		{Name: "aethergrok.exe", DownloadURL: "https://example.com/aethergrok.exe"},
		{Name: "AetherGrok-Setup.exe.sha256", DownloadURL: "https://example.com/AetherGrok-Setup.exe.sha256"},
	}
	matched := MatchPlatformAsset(assetsWithSetupAndZip, "windows", "amd64")
	if matched == nil || matched.Name != "AetherGrok-Setup.exe" {
		t.Fatalf("expected 'AetherGrok-Setup.exe', got %v", matched)
	}

	// Scenario 2: Setup absent -> Fallback to portable .zip
	assetsWithZipAndExe := []ReleaseAsset{
		{Name: "aethergrok-windows-amd64.zip", DownloadURL: "https://example.com/aethergrok-windows-amd64.zip"},
		{Name: "aethergrok.exe", DownloadURL: "https://example.com/aethergrok.exe"},
		{Name: "aethergrok-windows-amd64.zip.sha256", DownloadURL: "https://example.com/aethergrok-windows-amd64.zip.sha256"},
	}
	matched = MatchPlatformAsset(assetsWithZipAndExe, "windows", "amd64")
	if matched == nil || matched.Name != "aethergrok-windows-amd64.zip" {
		t.Fatalf("expected 'aethergrok-windows-amd64.zip', got %v", matched)
	}

	// Scenario 3: Setup and Zip absent -> Fallback to standalone executable (.exe)
	assetsWithOnlyExe := []ReleaseAsset{
		{Name: "aethergrok.exe", DownloadURL: "https://example.com/aethergrok.exe"},
		{Name: "aethergrok-darwin", DownloadURL: "https://example.com/aethergrok-darwin"},
		{Name: "aethergrok-linux", DownloadURL: "https://example.com/aethergrok-linux"},
	}
	matched = MatchPlatformAsset(assetsWithOnlyExe, "windows", "amd64")
	if matched == nil || matched.Name != "aethergrok.exe" {
		t.Fatalf("expected 'aethergrok.exe', got %v", matched)
	}

	// Scenario 4: Windows arm64 setup priority over zip and exe
	arm64Assets := []ReleaseAsset{
		{Name: "aethergrok-windows-arm64.exe", DownloadURL: "https://example.com/aethergrok-windows-arm64.exe"},
		{Name: "aethergrok-windows-arm64.zip", DownloadURL: "https://example.com/aethergrok-windows-arm64.zip"},
		{Name: "AetherGrok-1.0.2-windows-arm64-setup.exe", DownloadURL: "https://example.com/AetherGrok-1.0.2-windows-arm64-setup.exe"},
	}
	matched = MatchPlatformAsset(arm64Assets, "windows", "arm64")
	if matched == nil || matched.Name != "AetherGrok-1.0.2-windows-arm64-setup.exe" {
		t.Fatalf("expected 'AetherGrok-1.0.2-windows-arm64-setup.exe', got %v", matched)
	}

	// Scenario 5: Windows arm64 fallback to zip
	arm64ZipAssets := []ReleaseAsset{
		{Name: "aethergrok-windows-arm64.exe", DownloadURL: "https://example.com/aethergrok-windows-arm64.exe"},
		{Name: "aethergrok-windows-arm64.zip", DownloadURL: "https://example.com/aethergrok-windows-arm64.zip"},
	}
	matched = MatchPlatformAsset(arm64ZipAssets, "windows", "arm64")
	if matched == nil || matched.Name != "aethergrok-windows-arm64.zip" {
		t.Fatalf("expected 'aethergrok-windows-arm64.zip', got %v", matched)
	}

	// Scenario 6: Windows arm64 fallback to standalone exe
	arm64ExeAssets := []ReleaseAsset{
		{Name: "aethergrok-windows-arm64.exe", DownloadURL: "https://example.com/aethergrok-windows-arm64.exe"},
		{Name: "aethergrok-darwin-arm64", DownloadURL: "https://example.com/aethergrok-darwin-arm64"},
	}
	matched = MatchPlatformAsset(arm64ExeAssets, "windows", "arm64")
	if matched == nil || matched.Name != "aethergrok-windows-arm64.exe" {
		t.Fatalf("expected 'aethergrok-windows-arm64.exe', got %v", matched)
	}

	// Scenario 7: Canonical changelog keys
	canonicalAssets := []ReleaseAsset{
		{Name: "windows_x64_setup", DownloadURL: "https://example.com/windows_x64_setup"},
		{Name: "windows_arm64_setup", DownloadURL: "https://example.com/windows_arm64_setup"},
		{Name: "macos_arm64_dmg", DownloadURL: "https://example.com/macos_arm64_dmg"},
		{Name: "macos_x64_dmg", DownloadURL: "https://example.com/macos_x64_dmg"},
	}
	matchedWinX64 := MatchPlatformAsset(canonicalAssets, "windows", "amd64")
	if matchedWinX64 == nil || matchedWinX64.Name != "windows_x64_setup" {
		t.Fatalf("expected 'windows_x64_setup', got %v", matchedWinX64)
	}
	matchedWinArm64 := MatchPlatformAsset(canonicalAssets, "windows", "arm64")
	if matchedWinArm64 == nil || matchedWinArm64.Name != "windows_arm64_setup" {
		t.Fatalf("expected 'windows_arm64_setup', got %v", matchedWinArm64)
	}
}

func TestFindChecksumAsset(t *testing.T) {
	assets := []ReleaseAsset{
		{Name: "AetherGrok-Setup.exe", DownloadURL: "https://example.com/AetherGrok-Setup.exe"},
		{Name: "AetherGrok-Setup.exe.sha256", DownloadURL: "https://example.com/AetherGrok-Setup.exe.sha256"},
		{Name: "aethergrok-windows-amd64.zip", DownloadURL: "https://example.com/aethergrok-windows-amd64.zip"},
		{Name: "aethergrok-windows-amd64.zip.sha256", DownloadURL: "https://example.com/aethergrok-windows-amd64.zip.sha256"},
		{Name: "AetherGrok-1.0.1-macOS-arm64.dmg", DownloadURL: "https://example.com/AetherGrok-1.0.1-macOS-arm64.dmg"},
		{Name: "AetherGrok-1.0.1-macOS-arm64.dmg.sha256sum", DownloadURL: "https://example.com/AetherGrok-1.0.1-macOS-arm64.dmg.sha256sum"},
		{Name: "SHA256SUMS.txt", DownloadURL: "https://example.com/SHA256SUMS.txt"},
	}

	// 1. Target setup exe
	targetSetup := &ReleaseAsset{Name: "AetherGrok-Setup.exe"}
	shaAsset := FindChecksumAsset(assets, targetSetup)
	if shaAsset == nil || shaAsset.Name != "AetherGrok-Setup.exe.sha256" {
		t.Errorf("expected AetherGrok-Setup.exe.sha256, got %v", shaAsset)
	}

	// 2. Target zip
	targetZip := &ReleaseAsset{Name: "aethergrok-windows-amd64.zip"}
	shaZipAsset := FindChecksumAsset(assets, targetZip)
	if shaZipAsset == nil || shaZipAsset.Name != "aethergrok-windows-amd64.zip.sha256" {
		t.Errorf("expected aethergrok-windows-amd64.zip.sha256, got %v", shaZipAsset)
	}

	// 3. Target dmg with .sha256sum extension
	targetDMG := &ReleaseAsset{Name: "AetherGrok-1.0.1-macOS-arm64.dmg"}
	shaDMGAsset := FindChecksumAsset(assets, targetDMG)
	if shaDMGAsset == nil || shaDMGAsset.Name != "AetherGrok-1.0.1-macOS-arm64.dmg.sha256sum" {
		t.Errorf("expected AetherGrok-1.0.1-macOS-arm64.dmg.sha256sum, got %v", shaDMGAsset)
	}

	// 4. Target without specific checksum -> matches generic SHA256SUMS.txt
	targetGeneric := &ReleaseAsset{Name: "unknown-package.tar.gz"}
	shaGeneric := FindChecksumAsset(assets, targetGeneric)
	if shaGeneric == nil || shaGeneric.Name != "SHA256SUMS.txt" {
		t.Errorf("expected SHA256SUMS.txt, got %v", shaGeneric)
	}

	// 5. Assets without any checksum file
	noShaAssets := []ReleaseAsset{
		{Name: "AetherGrok-Setup.exe", DownloadURL: "https://example.com/AetherGrok-Setup.exe"},
	}
	if sha := FindChecksumAsset(noShaAssets, targetSetup); sha != nil {
		t.Errorf("expected nil for no checksum assets, got %v", sha)
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
