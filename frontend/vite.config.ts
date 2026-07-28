import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// dev 代理：前端跑在 5173，浏览器对 /api/* 的请求由 Vite 转发到后端 8080。
// 这样前端发的是"同源"请求 /api/health，绕开跨域(CORS)。
// 生产环境（前后端不同源）再由后端加 CORS 头，那是后面阶段的事。
export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {
      '/api': 'http://localhost:8080',
    },
  },
})
