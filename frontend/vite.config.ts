import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import tailwindcss from '@tailwindcss/vite';

export default defineConfig({
  plugins: [
    react({
      babel: {
        plugins: [
          'babel-plugin-transform-typescript-metadata',
          ['@babel/plugin-proposal-decorators', { legacy: true }],
        ],
      },
    }),
    tailwindcss(),
  ],
  server: {
    proxy: {
      '/admin': { target: 'http://localhost:8080' },
      '/v1': { target: 'http://localhost:8080' },
    },
  },
});
