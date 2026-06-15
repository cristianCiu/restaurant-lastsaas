import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'
import { VitePWA } from 'vite-plugin-pwa'

export default defineConfig({
  plugins: [
    react(),
    tailwindcss(),
    VitePWA({
      registerType: 'autoUpdate',
      includeAssets: ['favicon.ico'],
      manifest: {
        name: 'StockCount',
        short_name: 'StockCount',
        description: 'Daily restaurant stock tracking and forecasting',
        theme_color: '#0f172a',
        background_color: '#0f172a',
        display: 'standalone',
        icons: [
          { src: '/icon-192.png', sizes: '192x192', type: 'image/png' },
          { src: '/icon-512.png', sizes: '512x512', type: 'image/png' },
        ],
      },
      workbox: {
        globPatterns: ['**/*.{js,css,html,ico,png}'],
        runtimeCaching: [
          {
            urlPattern: /^\/api\/inventory\/stock-items/,
            handler: 'CacheFirst',
            options: {
              cacheName: 'stock-items-cache',
              expiration: { maxEntries: 10, maxAgeSeconds: 24 * 60 * 60 },
            },
          },
        ],
      },
    }),
  ],
  server: {
    port: parseInt(process.env.VITE_PORT || '4280'),
    proxy: {
      '/api': {
        target: process.env.VITE_API_URL || 'http://localhost:4290',
        changeOrigin: true,
      },
    },
  },
})
