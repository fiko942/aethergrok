export interface PackageInfo {
  label: string;
  osName: 'macOS' | 'Windows' | 'Linux' | 'Other';
  archName: 'Apple Silicon' | 'Intel x64' | 'x64' | 'ARM64' | 'Universal';
  kind: 'DMG Installer' | 'Setup Installer' | 'Portable Archive' | 'Binary';
  extension: string;
}

export function parsePackageInfo(rawName: string): PackageInfo {
  const lower = rawName.toLowerCase();

  let osName: PackageInfo['osName'] = 'Other';
  if (lower.includes('mac') || lower.includes('darwin')) {
    osName = 'macOS';
  } else if (lower.includes('win')) {
    osName = 'Windows';
  } else if (lower.includes('linux')) {
    osName = 'Linux';
  }

  let archName: PackageInfo['archName'] = 'x64';
  if (lower.includes('arm64') || lower.includes('aarch64')) {
    archName = osName === 'macOS' ? 'Apple Silicon' : 'ARM64';
  } else if (lower.includes('x64') || lower.includes('amd64') || lower.includes('x86_64')) {
    archName = osName === 'macOS' ? 'Intel x64' : 'x64';
  }

  let kind: PackageInfo['kind'] = 'Binary';
  let extension = '';
  if (lower.endsWith('.dmg') || lower.includes('_dmg')) {
    kind = 'DMG Installer';
    extension = '.dmg';
  } else if (lower.includes('setup') || lower.endsWith('.exe')) {
    kind = 'Setup Installer';
    extension = '.exe';
  } else if (lower.includes('portable') || lower.endsWith('.zip') || lower.endsWith('.tar.gz')) {
    kind = 'Portable Archive';
    extension = lower.includes('.tar.gz') ? '.tar.gz' : '.zip';
  }

  let label = `${osName} ${archName} (${kind})`;
  if (osName === 'macOS' && kind === 'DMG Installer') {
    label = `macOS ${archName} (DMG)`;
  } else if (osName === 'Windows' && kind === 'Setup Installer') {
    label = `Windows ${archName} Setup`;
  } else if (kind === 'Portable Archive') {
    label = `${osName} ${archName} Portable`;
  }

  return {
    label,
    osName,
    archName,
    kind,
    extension
  };
}
