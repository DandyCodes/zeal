import { defineConfig } from "vite"

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [],
  server: {
    proxy: {
      "/api": {
        target: "http://127.0.0.1:3975",
        changeOrigin: true,
        rewrite: path => path.replace(/^\/api/, ''),
      }
    },
    host: "127.0.0.1",
    port: 5173,
    strictPort: true
  }
})
