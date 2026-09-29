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

export const githubRepo = 'https://github.com/fiko942/aethergrok';

export interface ReleaseAsset {
  name: string;
  size: number;
  browser_download_url: string;
}

export interface LatestReleaseInfo {
  tagName: string;
  name: string;
  publishedAt: string;
  assets: ReleaseAsset[];
}

export async function fetchLiveLatestRelease(): Promise<LatestReleaseInfo | null> {
  try {
    const res = await fetch('https://api.github.com/repos/fiko942/aethergrok/releases/latest');
    if (!res.ok) return null;
    const data = await res.json();
    return {
      tagName: data.tag_name || 'v1.0.5',
      name: data.name || 'v1.0.5',
      publishedAt: data.published_at || '',
      assets: (data.assets || []).map((a: any) => ({
        name: a.name,
        size: a.size,
        browser_download_url: a.browser_download_url,
      })),
    };
  } catch (e) {
    return null;
  }
}
