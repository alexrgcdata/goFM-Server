import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  server: {
    // A fixed origin makes the matching server-side CORS setting predictable.
    port: 5173,
    strictPort: true,
  },
})
