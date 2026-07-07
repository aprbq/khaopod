import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'
import { fileURLToPath, URL } from 'node:url'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react(), tailwindcss()],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
  server: {
    port: 5173,
    // proxy ให้ browser คุยกับ backend แบบ same-origin — เลี่ยง CORS และให้ httpOnly cookie ทำงาน
    // ใช้ 127.0.0.1 (ไม่ใช่ localhost) เพราะ Node จะรีโซลฟ์ localhost เป็น ::1 (IPv6) ก่อน
    // แต่ Go server bind แค่ IPv4 → ECONNREFUSED; ล็อกเป็น IPv4 ตรง ๆ กันปัญหานี้
    proxy: {
      '/v1': 'http://127.0.0.1:8080',
      '/images': 'http://127.0.0.1:8080', // รูปสินค้า static จาก backend
      '/uploads': 'http://127.0.0.1:8080', // ไฟล์ที่ผู้ใช้อัปโหลด (รูปโปรไฟล์)
    },
  },
})
