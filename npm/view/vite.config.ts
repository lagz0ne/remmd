import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import { remmdPlugin } from './src/vite-plugin-remmd'

export default defineConfig({
  plugins: [react(), remmdPlugin()],
  server: {
    // No proxy needed — Go reverse-proxies to us
    // NATS WS is on Go's mux at /nats
  },
})
