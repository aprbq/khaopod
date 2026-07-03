import { useEffect } from 'react'

const SUFFIX = 'Khaopod Shop'

// ตั้งชื่อแท็บเบราว์เซอร์ต่อหน้า — ส่ง title มา จะได้ "<title> - Khaopod Shop"
// ไม่ส่ง (undefined) = ใช้แค่ "Khaopod Shop" (หน้าแรก)
export function useDocumentTitle(title?: string): void {
  useEffect(() => {
    document.title = title ? `${title} - ${SUFFIX}` : SUFFIX
  }, [title])
}
