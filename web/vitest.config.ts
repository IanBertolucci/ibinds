import { defineConfig, mergeConfig } from 'vitest/config';
import viteConfig from './vite.config.ts';

export default mergeConfig(
  viteConfig,
  defineConfig({
    // Svelte's package exports pick the server build unless the browser
    // condition is selected; component tests need the client runtime.
    resolve: {
      conditions: ['browser'],
    },
    test: {
      environment: 'happy-dom',
      include: ['src/**/*.test.ts'],
    },
  }),
);
