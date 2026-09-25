import { defineConfig } from 'vite';
import { svelte } from '@sveltejs/vite-plugin-svelte';
import path from 'path';
import pkg from './package.json';

export default defineConfig({
  plugins: [svelte()],
  define: {
    __APP_VERSION__: JSON.stringify(process.env.VITE_APP_VERSION || process.env.npm_package_version || pkg.version || '1.0.0'),
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
