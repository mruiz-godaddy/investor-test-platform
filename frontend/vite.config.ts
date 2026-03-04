import { defineConfig, type Plugin } from 'vite';
import react from '@vitejs/plugin-react';
import tailwindcss from '@tailwindcss/vite';

function requestLogger(): Plugin {
  return {
    name: 'request-logger',
    configureServer(server) {
      server.middlewares.use((req, res, next) => {
        const start = Date.now();
        const orig = res.end.bind(res) as typeof res.end;
        res.end = function (...args: Parameters<typeof res.end>) {
          const ms = Date.now() - start;
          const status = res.statusCode;
          const color = status < 400 ? '\x1b[32m' : '\x1b[31m';
          console.log(
            `${color}${req.method} ${req.url} → ${status}\x1b[0m (${ms}ms)`,
          );
          return orig(...args);
        } as typeof res.end;
        next();
      });
    },
  };
}

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
    requestLogger(),
  ],
  server: {
    proxy: {
      '/admin': { target: 'http://localhost:8080' },
      '/v1': { target: 'http://localhost:8080' },
    },
  },
});
