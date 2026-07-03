import { Outlet } from 'react-router-dom'
// import { AnnouncementBar } from '@/components/AnnouncementBar'
import { Footer } from '@/components/Footer'
import { Navbar } from '@/components/Navbar'

// โครงหน้าร้าน: แถบประกาศ + navbar + เนื้อหา + footer
export function StoreLayout() {
  return (
    <div className="flex min-h-screen flex-col">
      {/* <AnnouncementBar /> */}
      <Navbar />
      <main className="flex-1">
        <Outlet />
      </main>
      <Footer />
    </div>
  )
}
