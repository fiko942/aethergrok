# Windows Setup Installer Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Provide an automated, native Windows NSIS Setup Installer (`AetherGrok-<version>-windows-<arch>-setup.exe`) alongside the portable zip archive across GitHub Actions CI/CD and release pipelines.

**Architecture:** Configure NSIS installer templates in `build/windows/installer/` with custom branding (AetherGrok, desktop & start menu shortcuts, uninstaller), fix the PATH environment resolution for `makensis` on Windows GitHub Actions runner, and ensure proper naming, SHA256 checksumming, and GitHub Release asset publishing.

**Tech Stack:** Wails v2, NSIS (Nullsoft Scriptable Install System), GitHub Actions CI/CD, PowerShell, Node.js release script.

## Global Constraints

- Installer output naming: `AetherGrok-<version>-windows-<arch>-setup.exe` (e.g. `AetherGrok-1.0.1-windows-amd64-setup.exe` and `AetherGrok-1.0.1-windows-arm64-setup.exe`).
- Portable output naming: `AetherGrok-<version>-windows-<arch>-portable.zip`.
- Application Name: `AetherGrok`.
- Company Name: `AetherGrok` / `fiko942`.
- Install Scope: `user` or standard Program Files, creates Desktop shortcut and Start Menu shortcut with `icon.ico`.
- Uninstaller included and properly registered in Windows registry (`Add/Remove Programs`).

---

### Task 1: Configure Custom NSIS Installer Template and Metadata

**Files:**
- Create: `build/windows/installer/project.nsi`
- Modify: `wails.json`
- Modify: `build/windows/info.json`

**Interfaces:**
- Consumes: `wails.json` metadata (`name`, `version`, `author`), `build/windows/icon.ico`.
- Produces: `build/windows/installer/project.nsi` configured for AetherGrok setup packaging.

- [ ] **Step 1: Configure `wails.json` with info section**

Ensure `wails.json` includes `info` metadata with `productName`, `companyName`, and `copyright`:
```json
{
  "$schema": "https://wails.io/schemas/config.v2.json",
  "name": "aethergrok",
  "outputfilename": "aethergrok",
  "frontend:install": "pnpm install",
  "frontend:build": "pnpm build",
  "frontend:dev:watcher": "pnpm dev",
  "frontend:dev:serverUrl": "auto",
  "author": {
    "name": "fiko942",
    "email": "tobellord@gmail.com"
  },
  "info": {
    "companyName": "AetherGrok",
    "productName": "AetherGrok",
    "productVersion": "1.0.1",
    "copyright": "Copyright (c) 2026 Wiji Fiko Teren",
    "comments": "AetherGrok Desktop Studio"
  },
  "version": "1.0.1"
}
```

- [ ] **Step 2: Customize `build/windows/installer/project.nsi`**

Ensure `build/windows/installer/project.nsi` defines:
- `INFO_PROJECTNAME "aethergrok"`
- `INFO_COMPANYNAME "AetherGrok"`
- `INFO_PRODUCTNAME "AetherGrok"`
- Icon paths referencing `..\icon.ico`
- Desktop & Start menu shortcuts
- Complete uninstaller with registry registration

- [ ] **Step 3: Verify local files and commit**

```bash
git add wails.json build/windows/installer/project.nsi
git commit -m "feat(windows): configure NSIS installer template and application metadata"
```

---

### Task 2: Fix NSIS PATH and Installer Handling in GitHub Actions Workflow

**Files:**
- Modify: `.github/workflows/desktop-release.yml`

**Interfaces:**
- Consumes: Chocolatey NSIS installation on `windows-latest` runner.
- Produces: `C:\Program Files (x86)\NSIS` added to `$env:GITHUB_PATH` so `makensis` is found and executed during `wails build -nsis`.

- [ ] **Step 1: Update `.github/workflows/desktop-release.yml`**

Add NSIS binary directory to `GITHUB_PATH` after `choco install nsis`:
```yaml
      - name: Install NSIS
        shell: pwsh
        run: |
          choco install nsis -y
          $nsisPath = "C:\Program Files (x86)\NSIS"
          if (Test-Path "C:\Program Files\NSIS") { $nsisPath = "C:\Program Files\NSIS" }
          Write-Host "Adding NSIS to PATH: $nsisPath"
          echo "$nsisPath" | Out-File -FilePath $env:GITHUB_PATH -Encoding utf8 -Append
```

- [ ] **Step 2: Ensure Installer Rename and Packaging Match Release Pattern**

In `Build Windows Executable & Installer (${{ matrix.arch }})` step:
```powershell
          # Build executable and NSIS installer
          wails build -platform "windows/${{ matrix.arch }}" -nsis -clean -ldflags "-s -w"
          
          # Create Portable Zip Archive
          $TargetExe = "build/bin/aethergrok.exe"
          if (Test-Path "build/bin/AetherGrok.exe") { $TargetExe = "build/bin/AetherGrok.exe" }
          
          $PortableName = "AetherGrok-${Version}-windows-${{ matrix.arch }}-portable.zip"
          Compress-Archive -Path $TargetExe -DestinationPath "build/bin/$PortableName" -Force
          
          # Standardize Installer Name if exists
          $Installer = Get-Item "build/bin/*installer*.exe" -ErrorAction SilentlyContinue
          if ($Installer) {
            $StandardInstallerName = "AetherGrok-${Version}-windows-${{ matrix.arch }}-setup.exe"
            Rename-Item -Path $Installer.FullName -NewName $StandardInstallerName -Force
          }
```

- [ ] **Step 3: Commit workflow improvements**

```bash
git add .github/workflows/desktop-release.yml
git commit -m "ci(windows): ensure makensis is on PATH and output standard setup installer"
```

---

### Task 3: Verify Release Script and GitHub Release Table

**Files:**
- Modify: `scripts/release.mjs`

**Interfaces:**
- Consumes: Release artifacts `AetherGrok-${version}-windows-${arch}-setup.exe` and `AetherGrok-${version}-windows-${arch}-portable.zip`.
- Produces: GitHub Release markdown description with download links for all installer and portable variants.

- [ ] **Step 1: Check `scripts/release.mjs` artifact mappings**

Verify that `scripts/release.mjs` defines:
- `windows_x64_setup`: `AetherGrok-${nextVersion}-windows-amd64-setup.exe`
- `windows_x64_portable`: `AetherGrok-${nextVersion}-windows-amd64-portable.zip`
- `windows_arm64_setup`: `AetherGrok-${nextVersion}-windows-arm64-setup.exe`
- `windows_arm64_portable`: `AetherGrok-${nextVersion}-windows-arm64-portable.zip`

- [ ] **Step 2: Commit and push changes**

```bash
git add scripts/release.mjs
git commit -m "feat(release): verify full Windows setup installer and portable zip download matrix"
git push origin main
```
