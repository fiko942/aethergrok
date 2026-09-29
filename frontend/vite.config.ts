import { defineConfig } from 'vite';
import { svelte } from '@sveltejs/vite-plugin-svelte';
import path from 'path';
import pkg from './package.json';
import wailsConfig from '../wails.json';

const appVersion = process.env.VITE_APP_VERSION || 
                   process.env.npm_package_version || 
                   pkg.version || 
                   wailsConfig?.info?.productVersion || 
                   wailsConfig?.version || 
                   '1.0.5';

export default defineConfig({
  plugins: [svelte()],
  define: {
    __APP_VERSION__: JSON.stringify(appVersion),
  },
  resolve: {
    alias: {
      $lib: path.resolve(__dirname, './src/lib')
    }
  },
  server: {
    port: 5173,
    strictPort: true
  }
});
