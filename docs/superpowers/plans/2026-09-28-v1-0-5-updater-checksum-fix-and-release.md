# Plan: v1.0.5 Release - Auto-Updater SHA256 Checksum Matching & Multi-Arch Asset Alignment

**Date:** 2026-09-28  
**Status:** Completed & Shipped  
**Release Tag:** `v1.0.5`  
**Repository:** `https://github.com/fiko942/aethergrok`

---

## 1. Overview & Context

During in-app auto-update testing on Apple Silicon (macOS ARM64) upgrading from v1.0.0/v1.0.3 to v1.0.4, the in-app updater failed with an integrity error:
```text
integrity check failed: checksum mismatch: expected
'ab89cff1dfa65128aba04dfb7a983d4bdc6145a2ee185283e0d70e7a72e50e52 aethergrok-1.0.4-macos-amd64.dmg ',
got 'e25cda3e0435c61683c245f2445040a0bc3bbb7e109caa2bf85bd812e94818c4'
```

### Root Cause
In `frontend/src/lib/stores/updater.svelte.ts`, the frontend extracted the checksum URL by searching the release assets with `Array.prototype.find()` on `.sha256`:
```typescript
const shaAsset = this.latestRelease.assets.find(a => 
  a.name.endsWith('.sha256') || a.name.endsWith('.sha256sum') || a.name.includes('checksum')
);
```
Since GitHub Releases asset order placed `AetherGrok-1.0.4-macOS-amd64.dmg.sha256` first alphabetically, Apple Silicon hosts downloading `AetherGrok-1.0.4-macOS-arm64.dmg` received the AMD64 checksum file URL. The Go backend (`pkg/updater/installer.go`) then compared the ARM64 binary against the AMD64 hash, resulting in a checksum mismatch abort.

---

## 2. Engineering Changes

### 1. Hierarchical Checksum Resolution (`frontend/src/lib/stores/updater.svelte.ts`)
Updated `startDownloadAndInstall()` to match checksum files with three tiered strategies:
1. **Exact Target File Extension Match**: Check for `<targetFilename>.sha256` or `<targetFilename>.sha256sum`.
2. **Platform & Arch Substring Matching**: Filter candidate checksum files by OS format (`.dmg`, `.exe`, `.zip`) and target architecture (`arm64`/`aarch64` vs `amd64`/`x64`/`intel`).
3. **Safe Fallback**: Use generic checksum only when no specific file or architecture matches.

### 2. Workspace & Web Distribution Packaging
- Added `web` to `pnpm-workspace.yaml` packages.
- Updated `pnpm-lock.yaml` to ensure clean workspace installations.
- Rebuilt production static assets in `web/dist/` aligning `displayVersion` to `v1.0.5`.

### 3. Release Automation & Artifact Verification
- Ran pre-flight verification across TypeScript, Vitest (2000+ tests passing), Svelte production build, and Go backend packages.
- Published release tag `v1.0.5` with release title:
  `v1.0.5 - AetherGrok 1.0.5 - Auto-Updater SHA256 Checksum Matching Fix & Web Updates`
- GitHub Actions CI/CD completed all jobs (`Build macOS (arm64)`, `Build macOS (amd64)`, `Build Windows (amd64)`, `Build Windows (arm64)`), uploading all 14 binary packages and checksum files.

---

## 3. Verification & Evidence
- **Vitest Suite**: 2,000+ unit and DOM tests passing.
- **Go Tests**: `go test ./pkg/updater/...` passing.
- **Release Assets**: 14 assets verified on `https://github.com/fiko942/aethergrok/releases/tag/v1.0.5`.
