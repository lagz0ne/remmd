import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'
import { remmdPlugin } from './src/vite-plugin-remmd'

export default defineConfig({
  plugins: [react(), tailwindcss(), remmdPlugin()],
  define: {
    __BUILD_VERSION__: JSON.stringify(
      `${new Date().toISOString().slice(0, 16)} | ${process.env.npm_package_version || 'dev'}`
    ),
  },
  server: {},
})
