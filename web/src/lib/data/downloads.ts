export interface DownloadPackage {
  platform: 'macOS' | 'Windows';
  arch: string;
  type: string;
  filename: string;
  size: string;
  url: string;
  checksumUrl?: string;
  recommended?: boolean;
}

export const releaseVersion = '1.0.3';
export const releaseDate = 'September 2026';
export const githubRepo = 'https://github.com/fiko942/aethergrok';

export const downloadsData: DownloadPackage[] = [
  {
    platform: 'macOS',
    arch: 'Apple Silicon (M1 / M2 / M3 / M4)',
    type: 'Finder Styled .dmg',
    filename: `AetherGrok-${releaseVersion}-macOS-arm64.dmg`,
    size: '6.5 MB',
    url: `${githubRepo}/releases/download/v${releaseVersion}/AetherGrok-${releaseVersion}-macOS-arm64.dmg`,
    checksumUrl: `${githubRepo}/releases/download/v${releaseVersion}/AetherGrok-${releaseVersion}-macOS-arm64.dmg.sha256`,
    recommended: true,
  },
  {
    platform: 'macOS',
    arch: 'Intel x86_64',
    type: 'Finder Styled .dmg',
    filename: `AetherGrok-${releaseVersion}-macOS-amd64.dmg`,
    size: '7.0 MB',
    url: `${githubRepo}/releases/download/v${releaseVersion}/AetherGrok-${releaseVersion}-macOS-amd64.dmg`,
    checksumUrl: `${githubRepo}/releases/download/v${releaseVersion}/AetherGrok-${releaseVersion}-macOS-amd64.dmg.sha256`,
  },
  {
    platform: 'Windows',
    arch: 'x64 (64-bit AMD/Intel)',
    type: 'Setup Installer .exe',
    filename: `AetherGrok-${releaseVersion}-windows-amd64-setup.exe`,
    size: '8.2 MB',
    url: `${githubRepo}/releases/download/v${releaseVersion}/AetherGrok-${releaseVersion}-windows-amd64-setup.exe`,
    recommended: true,
  },
  {
    platform: 'Windows',
    arch: 'x64 (64-bit AMD/Intel)',
    type: 'Portable Standalone .zip',
    filename: `AetherGrok-${releaseVersion}-windows-amd64-portable.zip`,
    size: '7.9 MB',
    url: `${githubRepo}/releases/download/v${releaseVersion}/AetherGrok-${releaseVersion}-windows-amd64-portable.zip`,
  },
  {
    platform: 'Windows',
    arch: 'ARM64 (Snapdragon X / ARM)',
    type: 'Setup Installer .exe',
    filename: `AetherGrok-${releaseVersion}-windows-arm64-setup.exe`,
    size: '7.8 MB',
    url: `${githubRepo}/releases/download/v${releaseVersion}/AetherGrok-${releaseVersion}-windows-arm64-setup.exe`,
  },
  {
    platform: 'Windows',
    arch: 'ARM64 (Snapdragon X / ARM)',
    type: 'Portable Standalone .zip',
    filename: `AetherGrok-${releaseVersion}-windows-arm64-portable.zip`,
    size: '7.5 MB',
    url: `${githubRepo}/releases/download/v${releaseVersion}/AetherGrok-${releaseVersion}-windows-arm64-portable.zip`,
  },
];
