package updater

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// ReleaseAsset represents a downloadable file from a release
type ReleaseAsset struct {
	Name        string `json:"name"`
	Size        int64  `json:"size"`
	DownloadURL string `json:"downloadUrl"`
	ContentType string `json:"contentType"`
}

// ReleaseInfo contains parsed details of a version release
type ReleaseInfo struct {
	Version     string         `json:"version"`
	TagName     string         `json:"tagName"`
	Title       string         `json:"title"`
	PublishedAt string         `json:"publishedAt"`
	Body        string         `json:"body"`
	Highlights  []string       `json:"highlights"`
	Assets      []ReleaseAsset `json:"assets"`
	IsLatest    bool           `json:"isLatest"`
}

// UpdateCheckResult is returned by CheckForUpdates to the frontend
type UpdateCheckResult struct {
	UpdateAvailable bool          `json:"updateAvailable"`
	CurrentVersion  string        `json:"currentVersion"`
	LatestVersion   string        `json:"latestVersion"`
	LatestRelease   *ReleaseInfo  `json:"latestRelease"`
	AllReleases     []ReleaseInfo `json:"allReleases"`
	PlatformAsset   *ReleaseAsset `json:"platformAsset"`
	CheckedAt       string        `json:"checkedAt"`
}

// gitHubAsset represents the JSON structure from GitHub Release API assets
type gitHubAsset struct {
	Name               string `json:"name"`
	Size               int64  `json:"size"`
	BrowserDownloadURL string `json:"browser_download_url"`
	ContentType        string `json:"content_type"`
}

// gitHubRelease represents the JSON structure from GitHub Release API
type gitHubRelease struct {
	TagName     string        `json:"tag_name"`
	Name        string        `json:"name"`
	Body        string        `json:"body"`
	PublishedAt string        `json:"published_at"`
	Draft       bool          `json:"draft"`
	Prerelease  bool          `json:"prerelease"`
	Assets      []gitHubAsset `json:"assets"`
}

// changelogJSON represents the schema of local changelog.json
type changelogJSON struct {
	Latest   string `json:"latest"`
	Releases []struct {
		Version    string            `json:"version"`
		Tag        string            `json:"tag"`
		Date       string            `json:"date"`
		Title      string            `json:"title"`
		Highlights []string          `json:"highlights"`
		Downloads  map[string]string `json:"downloads"`
	} `json:"releases"`
}

// CleanVersion normalizes a semver string by trimming spaces and stripping 'v' prefix
func CleanVersion(v string) string {
	v = strings.TrimSpace(v)
	v = strings.TrimPrefix(v, "v")
	v = strings.TrimPrefix(v, "V")
	return v
}

// parseSemVerParts extracts [major, minor, patch] integers and optional prerelease suffix
func parseSemVerParts(v string) ([]int, string) {
	clean := CleanVersion(v)
	if clean == "" {
		return []int{0, 0, 0}, ""
	}

	prerelease := ""
	if idx := strings.Index(clean, "-"); idx != -1 {
		prerelease = clean[idx+1:]
		clean = clean[:idx]
	}

	parts := strings.Split(clean, ".")
	nums := make([]int, 3)
	for i := 0; i < len(parts) && i < 3; i++ {
		numStr := parts[i]
		// extract leading digits only if any trailing metadata exists
		var digits []rune
		for _, r := range numStr {
			if r >= '0' && r <= '9' {
				digits = append(digits, r)
			} else {
				break
			}
		}
		if len(digits) > 0 {
			n, err := strconv.Atoi(string(digits))
			if err == nil {
				nums[i] = n
			}
		}
	}
	return nums, prerelease
}

// CompareSemVer compares two semantic version strings.
// Returns 1 if v1 > v2, -1 if v1 < v2, and 0 if equal.
// Handles 'v' prefixes, major.minor.patch, and basic pre-releases.
func CompareSemVer(v1, v2 string) int {
	nums1, pre1 := parseSemVerParts(v1)
	nums2, pre2 := parseSemVerParts(v2)

	for i := 0; i < 3; i++ {
		if nums1[i] > nums2[i] {
			return 1
		}
		if nums1[i] < nums2[i] {
			return -1
		}
	}

	// If numeric parts are equal, handle prerelease tags
	// SemVer spec: a version without a prerelease tag has higher precedence than one with a prerelease tag.
	if pre1 == "" && pre2 != "" {
		return 1
	}
	if pre1 != "" && pre2 == "" {
		return -1
	}
	if pre1 != "" && pre2 != "" {
		if pre1 > pre2 {
			return 1
		}
		if pre1 < pre2 {
			return -1
		}
	}

	return 0
}

// scoreReleaseAsset evaluates how well an asset filename matches target OS and arch.
// Higher score indicates better match. A score <= 0 indicates not matching.
func scoreReleaseAsset(name string, isWindows, isDarwin, isARM64, isAMD64 bool) int {
	lower := strings.ToLower(name)

	// Exclude non-executable / checksum / metadata files
	ignoredExts := []string{
		".sha256", ".sha256sum", ".md5", ".sig", ".asc",
		".sha1", ".sha512", ".blockmap", ".txt", ".json", ".md", ".log",
	}
	for _, ext := range ignoredExts {
		if strings.HasSuffix(lower, ext) {
			return 0
		}
	}

	if isDarwin {
		// Reject assets clearly for other OSes
		if strings.Contains(lower, "win") || strings.Contains(lower, "windows") ||
			strings.Contains(lower, "linux") || strings.Contains(lower, "android") ||
			strings.HasSuffix(lower, ".exe") || strings.HasSuffix(lower, ".deb") ||
			strings.HasSuffix(lower, ".rpm") || strings.HasSuffix(lower, ".appimage") {
			return 0
		}

		isDMG := strings.HasSuffix(lower, ".dmg") || strings.Contains(lower, "dmg")
		isZip := strings.HasSuffix(lower, ".zip") || strings.Contains(lower, "zip") || strings.HasSuffix(lower, ".tar.gz")
		hasMacTag := strings.Contains(lower, "mac") || strings.Contains(lower, "darwin") || strings.Contains(lower, "osx") || strings.Contains(lower, "apple")

		if !isDMG && !isZip && !hasMacTag {
			return 0
		}

		hasArm64 := strings.Contains(lower, "arm64") || strings.Contains(lower, "aarch64") ||
			strings.Contains(lower, "apple silicon") || strings.Contains(lower, "applesilicon") ||
			strings.Contains(lower, "m1") || strings.Contains(lower, "m2") || strings.Contains(lower, "m3")
		hasAmd64 := strings.Contains(lower, "amd64") || strings.Contains(lower, "x64") ||
			strings.Contains(lower, "x86_64") || strings.Contains(lower, "intel")

		if isARM64 {
			if isDMG {
				if hasArm64 {
					return 350
				}
				if !hasAmd64 {
					return 200 // generic/universal DMG
				}
			} else if isZip {
				if hasArm64 {
					return 250
				}
				if !hasAmd64 {
					return 180
				}
			}
		} else if isAMD64 {
			if isDMG {
				if hasAmd64 {
					return 350
				}
				if !hasArm64 {
					return 200 // generic/universal DMG
				}
			} else if isZip {
				if hasAmd64 {
					return 250
				}
				if !hasArm64 {
					return 180
				}
			}
		}
		return 0
	}

	if isWindows {
		// Reject assets clearly for other OSes
		if (strings.Contains(lower, "darwin") || strings.Contains(lower, "macos") ||
			strings.Contains(lower, "mac-") || strings.Contains(lower, "mac_") ||
			strings.Contains(lower, "linux") || strings.Contains(lower, "android") ||
			strings.HasSuffix(lower, ".dmg") || strings.HasSuffix(lower, ".deb") ||
			strings.HasSuffix(lower, ".rpm") || strings.HasSuffix(lower, ".appimage")) &&
			!strings.Contains(lower, "windows") && !strings.Contains(lower, "win") {
			return 0
		}

		isExe := strings.HasSuffix(lower, ".exe") || strings.Contains(lower, "_exe") || strings.Contains(lower, ".exe.")
		isZip := strings.HasSuffix(lower, ".zip") || strings.Contains(lower, "_zip") || strings.HasSuffix(lower, ".7z")
		hasWinTag := strings.Contains(lower, "win") || strings.Contains(lower, "windows")
		isSetup := strings.Contains(lower, "setup") || strings.Contains(lower, "installer")

		// If it has no windows indicator, exe suffix, zip suffix, or setup/installer keyword, skip
		if !isExe && !isZip && !hasWinTag && !isSetup {
			return 0
		}

		hasArm64 := strings.Contains(lower, "arm64") || strings.Contains(lower, "aarch64") || strings.Contains(lower, "armv8")
		hasAmd64 := strings.Contains(lower, "amd64") || strings.Contains(lower, "x64") ||
			strings.Contains(lower, "x86_64") || strings.Contains(lower, "x86-64") || strings.Contains(lower, "intel")
		isGenericArch := !hasArm64 && !hasAmd64

		if isAMD64 {
			// ARM64 binaries should not match AMD64 Windows
			if hasArm64 && !hasAmd64 {
				return 0
			}

			// Priority 1: NSIS Setup / Installer (.exe)
			if isSetup || (isExe && strings.Contains(lower, "setup")) {
				if hasAmd64 {
					return 350 // e.g. AetherGrok-1.0.1-windows-x64-setup.exe, windows_x64_setup
				}
				if isGenericArch || hasWinTag {
					return 320 // e.g. AetherGrok-Setup.exe, setup.exe, installer.exe
				}
			}

			// Priority 2: Portable .zip
			if isZip {
				if hasAmd64 {
					return 250 // e.g. aethergrok-windows-amd64.zip, windows_x64_zip
				}
				if isGenericArch || hasWinTag {
					return 220 // e.g. aethergrok-windows.zip, aethergrok-portable.zip
				}
			}

			// Priority 3: Standalone executable (.exe)
			if isExe {
				if hasAmd64 {
					return 150 // e.g. aethergrok-windows-amd64.exe, windows_x64_exe
				}
				if isGenericArch || hasWinTag {
					return 120 // e.g. aethergrok.exe, aethergrok-windows.exe
				}
			}

			if hasAmd64 {
				return 50
			}
		} else if isARM64 {
			// Priority 1: NSIS Setup / Installer for ARM64
			if (isSetup || isExe) && hasArm64 && isSetup {
				return 350 // e.g. AetherGrok-1.0.1-windows-arm64-setup.exe, windows_arm64_setup
			}

			// Priority 2: Portable .zip for ARM64
			if isZip && hasArm64 {
				return 250 // e.g. aethergrok-windows-arm64.zip, windows_arm64_zip
			}

			// Priority 3: Standalone executable for ARM64
			if isExe && hasArm64 {
				return 150 // e.g. aethergrok-windows-arm64.exe, windows_arm64_exe
			}

			// Fallbacks if no ARM64 specific binary exists
			if isSetup && isGenericArch {
				return 80 // generic setup
			}
			if isZip && (isGenericArch || hasWinTag) {
				return 60
			}
			if isExe && (isGenericArch || hasWinTag) {
				return 40
			}
			if hasAmd64 && isSetup {
				return 20 // AMD64 emulation fallback
			}
		}
	}

	return 0
}

// MatchPlatformAsset finds the best matching release asset for the given OS and architecture.
// osName: "darwin" / "macos", "windows" / "win32"
// arch: "arm64" / "aarch64", "amd64" / "x64" / "x86_64"
func MatchPlatformAsset(assets []ReleaseAsset, osName, arch string) *ReleaseAsset {
	if len(assets) == 0 {
		return nil
	}

	normOS := strings.ToLower(osName)
	normArch := strings.ToLower(arch)

	isDarwin := normOS == "darwin" || normOS == "macos" || strings.Contains(normOS, "mac") || strings.Contains(normOS, "osx")
	isWindows := normOS == "windows" || normOS == "win32" || strings.Contains(normOS, "win")
	isARM64 := normArch == "arm64" || normArch == "aarch64" || strings.Contains(normArch, "arm")
	isAMD64 := normArch == "amd64" || normArch == "x64" || normArch == "x86_64" || normArch == "amd"

	var bestAsset *ReleaseAsset
	bestScore := 0

	for i := range assets {
		asset := &assets[i]
		score := scoreReleaseAsset(asset.Name, isWindows, isDarwin, isARM64, isAMD64)
		if score > bestScore {
			bestScore = score
			bestAsset = asset
		}
	}

	return bestAsset
}

// FindChecksumAsset searches a list of release assets for the checksum file (.sha256 / .sha256sum)
// that corresponds to the given target binary asset.
func FindChecksumAsset(assets []ReleaseAsset, targetAsset *ReleaseAsset) *ReleaseAsset {
	if targetAsset == nil || len(assets) == 0 {
		return nil
	}

	targetName := strings.ToLower(targetAsset.Name)
	targetExt := filepath.Ext(targetName)
	baseWithoutExt := strings.TrimSuffix(targetName, targetExt)

	// 1. Exact match: <targetName>.sha256 or <targetName>.sha256sum
	for i := range assets {
		a := &assets[i]
		lower := strings.ToLower(a.Name)
		if lower == targetName+".sha256" || lower == targetName+".sha256sum" {
			return a
		}
	}

	// 2. Base without ext match: <baseWithoutExt>.sha256 or <baseWithoutExt>.sha256sum
	for i := range assets {
		a := &assets[i]
		lower := strings.ToLower(a.Name)
		if lower == baseWithoutExt+".sha256" || lower == baseWithoutExt+".sha256sum" {
			return a
		}
	}

	// 3. Platform & arch heuristic match
	isARM64 := strings.Contains(targetName, "arm64") || strings.Contains(targetName, "aarch64")
	isAMD64 := strings.Contains(targetName, "amd64") || strings.Contains(targetName, "x64") || strings.Contains(targetName, "x86_64") || strings.Contains(targetName, "intel")
	isWin := strings.Contains(targetName, "win") || strings.Contains(targetName, "windows") || strings.HasSuffix(targetName, ".exe")
	isMac := strings.Contains(targetName, "mac") || strings.Contains(targetName, "darwin") || strings.HasSuffix(targetName, ".dmg")

	for i := range assets {
		a := &assets[i]
		lower := strings.ToLower(a.Name)
		if !strings.HasSuffix(lower, ".sha256") && !strings.HasSuffix(lower, ".sha256sum") && !strings.Contains(lower, "checksum") {
			continue
		}

		if isMac && (strings.Contains(lower, "mac") || strings.Contains(lower, "darwin") || strings.Contains(lower, "dmg")) {
			if isARM64 && (strings.Contains(lower, "arm64") || strings.Contains(lower, "aarch64")) {
				return a
			}
			if isAMD64 && (strings.Contains(lower, "amd64") || strings.Contains(lower, "x64") || strings.Contains(lower, "intel")) {
				return a
			}
		}

		if isWin && (strings.Contains(lower, "win") || strings.Contains(lower, "windows") || strings.Contains(lower, "exe") || strings.Contains(lower, "zip")) {
			if isARM64 && (strings.Contains(lower, "arm64") || strings.Contains(lower, "aarch64")) {
				return a
			}
			if isAMD64 && (strings.Contains(lower, "amd64") || strings.Contains(lower, "x64") || strings.Contains(lower, "intel")) {
				return a
			}
		}
	}

	// 4. Generic SHA256SUMS / checksums bundle file
	for i := range assets {
		a := &assets[i]
		lower := strings.ToLower(a.Name)
		if lower == "sha256sums" || lower == "sha256sums.txt" ||
			lower == "checksums" || lower == "checksums.txt" ||
			lower == "checksums.sha256" || lower == "release.sha256" ||
			strings.HasPrefix(lower, "sha256sums") || strings.HasPrefix(lower, "checksums") {
			return a
		}
	}

	return nil
}

// ExtractHighlightsFromMarkdown extracts bullet points from markdown body
func ExtractHighlightsFromMarkdown(body string) []string {
	var highlights []string
	lines := strings.Split(body, "\n")
	bulletRegex := regexp.MustCompile(`^[-*•]\s+(.*)$`)

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		matches := bulletRegex.FindStringSubmatch(trimmed)
		if len(matches) > 1 {
			bulletText := strings.TrimSpace(matches[1])
			if bulletText != "" {
				highlights = append(highlights, bulletText)
			}
		}
	}
	return highlights
}

// loadLocalChangelog attempts to find and parse changelog.json from the filesystem
func loadLocalChangelog() ([]ReleaseInfo, error) {
	searchPaths := []string{
		"changelog.json",
		filepath.Join("..", "changelog.json"),
	}

	if exePath, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exePath)
		searchPaths = append(searchPaths,
			filepath.Join(exeDir, "changelog.json"),
			filepath.Join(exeDir, "..", "Resources", "changelog.json"),
			filepath.Join(exeDir, "..", "..", "Resources", "changelog.json"),
		)
	}

	var data []byte
	var readErr error
	for _, p := range searchPaths {
		data, readErr = os.ReadFile(p)
		if readErr == nil && len(data) > 0 {
			break
		}
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("local changelog.json not found: %w", readErr)
	}

	var parsed changelogJSON
	if err := json.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("failed to parse changelog.json: %w", err)
	}

	var releases []ReleaseInfo
	for i, r := range parsed.Releases {
		var assets []ReleaseAsset
		for name, dlURL := range r.Downloads {
			assets = append(assets, ReleaseAsset{
				Name:        name,
				DownloadURL: dlURL,
				ContentType: "application/octet-stream",
			})
		}

		info := ReleaseInfo{
			Version:     r.Version,
			TagName:     r.Tag,
			Title:       r.Title,
			PublishedAt: r.Date,
			Highlights:  r.Highlights,
			Assets:      assets,
			IsLatest:    i == 0 || r.Version == parsed.Latest,
		}
		releases = append(releases, info)
	}

	return releases, nil
}

// FetchReleases retrieves release info from GitHub API with local fallback
func FetchReleases(repo string) ([]ReleaseInfo, error) {
	if repo == "" {
		repo = "fiko942/aethergrok"
	}

	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/releases?per_page=20", repo)
	req, err := http.NewRequest("GET", apiURL, nil)
	if err == nil {
		req.Header.Set("User-Agent", "AetherGrok-Desktop-Updater")
		req.Header.Set("Accept", "application/vnd.github.v3+json")

		client := &http.Client{
			Timeout: 10 * time.Second,
		}

		resp, httpErr := client.Do(req)
		if httpErr == nil && resp.StatusCode == http.StatusOK {
			defer resp.Body.Close()
			bodyBytes, readErr := io.ReadAll(resp.Body)
			if readErr == nil {
				var ghReleases []gitHubRelease
				if unmarshalErr := json.Unmarshal(bodyBytes, &ghReleases); unmarshalErr == nil && len(ghReleases) > 0 {
					var releases []ReleaseInfo
					for i, gh := range ghReleases {
						if gh.Draft {
							continue
						}

						var assets []ReleaseAsset
						for _, a := range gh.Assets {
							assets = append(assets, ReleaseAsset{
								Name:        a.Name,
								Size:        a.Size,
								DownloadURL: a.BrowserDownloadURL,
								ContentType: a.ContentType,
							})
						}

						ver := CleanVersion(gh.TagName)
						highlights := ExtractHighlightsFromMarkdown(gh.Body)

						title := gh.Name
						if title == "" {
							title = fmt.Sprintf("AetherGrok %s", ver)
						}

						releases = append(releases, ReleaseInfo{
							Version:     ver,
							TagName:     gh.TagName,
							Title:       title,
							PublishedAt: gh.PublishedAt,
							Body:        gh.Body,
							Highlights:  highlights,
							Assets:      assets,
							IsLatest:    i == 0,
						})
					}

					if len(releases) > 0 {
						return releases, nil
					}
				}
			}
		}
	}

	// Fallback to local changelog.json if GitHub API failed or was rate limited
	localReleases, localErr := loadLocalChangelog()
	if localErr == nil && len(localReleases) > 0 {
		return localReleases, nil
	}

	if err != nil {
		return nil, err
	}
	return nil, fmt.Errorf("failed to fetch releases from GitHub and local fallback changelog was not available")
}

// CheckForUpdates inspects whether a newer release is available compared to currentVersion
func CheckForUpdates(currentVersion, repo string, osName, arch string) (*UpdateCheckResult, error) {
	releases, err := FetchReleases(repo)
	if err != nil {
		return nil, err
	}

	if len(releases) == 0 {
		return &UpdateCheckResult{
			UpdateAvailable: false,
			CurrentVersion:  currentVersion,
			LatestVersion:   currentVersion,
			CheckedAt:       time.Now().UTC().Format(time.RFC3339),
		}, nil
	}

	latestRelease := &releases[0]
	// Find the release with the highest SemVer
	for i := range releases {
		if CompareSemVer(releases[i].Version, latestRelease.Version) > 0 {
			latestRelease = &releases[i]
		}
	}

	updateAvailable := CompareSemVer(latestRelease.Version, currentVersion) > 0
	platformAsset := MatchPlatformAsset(latestRelease.Assets, osName, arch)

	return &UpdateCheckResult{
		UpdateAvailable: updateAvailable,
		CurrentVersion:  currentVersion,
		LatestVersion:   latestRelease.Version,
		LatestRelease:   latestRelease,
		AllReleases:     releases,
		PlatformAsset:   platformAsset,
		CheckedAt:       time.Now().UTC().Format(time.RFC3339),
	}, nil
}
