# Multi-Platform Packaging, DMG Clean Layout & Windows Setup Installer

## Context & Objectives
AetherGrok Desktop Studio targets both macOS (Apple Silicon `arm64` and Intel `amd64`) and Windows (x64 `amd64` and `arm64`). Distributables require standard, reliable installers and portable archives without platform-specific visual glitches or missing build tools in CI/CD runners.

## macOS Packaging Architecture (`build-macos.sh`)
1. **Clean Native Finder Layout**:
   - Custom graphical backgrounds (`.background/background.png` / TIFF) were removed to avoid contrast issues in light/dark mode and unwanted whitespace.
   - Finder layout is configured via AppleScript without any custom background image:
     - Window bounds: `{320, 160, 860, 480}` (540x320 viewport)
     - Clean window view: `icon view`, `toolbar visible: false`, `statusbar visible: false`, `pathbar visible: false`
     - Icon settings: `arrangement: not arranged`, `icon size: 96`, `label position: bottom`, `text size: 12`
     - Symmetrical icon positions:
       - `AetherGrok.app`: `{140, 130}`
       - `/Applications` symlink: `{400, 130}`
2. **Compression & Read-Only Output**:
   - Staging directory converted via `hdiutil create -format UDRW -fs HFS+`.
   - Layout applied on mounted volume, unmounted cleanly (`hdiutil detach`).
   - Final read-only DMG converted via `hdiutil convert -format UDZO -imagekey zlib-level=9`.

## Windows NSIS Setup Installer Architecture
1. **NSIS Template & Assets (`build/windows/installer/`)**:
   - `project.nsi`: Standard NSIS script configured for AetherGrok (`INFO_PROJECTNAME: aethergrok`, `INFO_PRODUCTNAME: AetherGrok`, `INFO_COMPANYNAME: AetherGrok`).
   - Creates Desktop shortcut (`$DESKTOP\AetherGrok.lnk`) and Start Menu shortcut (`$SMPROGRAMS\AetherGrok.lnk`) pointing to `$INSTDIR\aethergrok.exe` with icon from `build/windows/icon.ico`.
   - Generates and registers full uninstaller in Windows Registry (`Software\Microsoft\Windows\CurrentVersion\Uninstall\aethergrokaethergrok`) for clean removal via Settings / Control Panel.
   - Includes WebView2 bootstrapper installer support (`wails.webview2runtime`).
2. **GitHub Actions Environment Fix (`.github/workflows/desktop-release.yml`)**:
   - On `windows-latest` runners, `choco install nsis -y` installs `makensis.exe` to `C:\Program Files (x86)\NSIS` or `C:\Program Files\NSIS`.
   - Critical fix: Chocolatey installation does not automatically add the NSIS folder to the current step's shell environment. The workflow explicitly appends the resolved NSIS path to `$env:GITHUB_PATH` and `$env:PATH`:
     ```powershell
     $nsisPath = "C:\Program Files (x86)\NSIS"
     if (Test-Path "C:\Program Files\NSIS") { $nsisPath = "C:\Program Files\NSIS" }
     echo "$nsisPath" | Out-File -FilePath $env:GITHUB_PATH -Encoding utf8 -Append
     ```
   - Build step verifies `wails build -platform "windows/${{ matrix.arch }}" -nsis -clean -ldflags "-s -w"` and standardizes installer output naming.

## Standardized Release Artifact Matrix (`scripts/release.mjs`)
The multi-platform release manager recognizes 6 canonical package artifacts:
1. `AetherGrok-${version}-macOS-arm64.dmg` (macOS Apple Silicon DMG)
2. `AetherGrok-${version}-macOS-amd64.dmg` (macOS Intel x64 DMG)
3. `AetherGrok-${version}-windows-amd64-setup.exe` (Windows x64 NSIS Installer)
4. `AetherGrok-${version}-windows-amd64-portable.zip` (Windows x64 Portable)
5. `AetherGrok-${version}-windows-arm64-setup.exe` (Windows ARM64 NSIS Installer)
6. `AetherGrok-${version}-windows-arm64-portable.zip` (Windows ARM64 Portable)

Each artifact is accompanied by an automated SHA-256 checksum file (`*.sha256`).
