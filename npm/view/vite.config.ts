import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'
import { remmdPlugin } from './src/vite-plugin-remmd'

export default defineConfig({
  plugins: [react(), tailwindcss(), remmdPlugin()],
  server: {
    // No proxy needed — Go reverse-proxies to us
    // NATS WS is on Go's mux at /nats
  },
})
