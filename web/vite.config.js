import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  // Lets this admin panel live in its own subfolder without replacing an
  // existing OpenBridge site's index.html.
  base: process.env.VITE_APP_BASE_PATH || '/',
  plugins: [react()],
  server: {
    // A fixed origin makes the matching server-side CORS setting predictable.
    port: 5173,
    strictPort: true,
  },
})
