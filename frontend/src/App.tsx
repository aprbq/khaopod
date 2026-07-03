import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom'
import { StoreLayout } from '@/components/StoreLayout'
import { ProtectedRoute } from '@/features/auth/ProtectedRoute'
import { CartPage } from '@/pages/CartPage'
import { HomePage } from '@/pages/HomePage'
import { LoginPage } from '@/pages/LoginPage'
import { ProductDetailPage } from '@/pages/ProductDetailPage'
import { ProfilePage } from '@/pages/ProfilePage'
import { ShopPage } from '@/pages/ShopPage'
import { VerifyOtpPage } from '@/pages/VerifyOtpPage'

export function App() {
  return (
    <BrowserRouter>
      <Routes>
        {/* หน้าร้าน (มี navbar/footer) */}
        <Route element={<StoreLayout />}>
          <Route path="/" element={<HomePage />} />
          <Route path="/shop" element={<ShopPage />} />
          <Route path="/product/:slug" element={<ProductDetailPage />} />
          <Route path="/cart" element={<CartPage />} />

          {/* ต้องล็อกอิน */}
          <Route element={<ProtectedRoute />}>
            <Route path="/account" element={<ProfilePage />} />
          </Route>
        </Route>

        {/* หน้า auth (เต็มจอ ไม่มี navbar) */}
        <Route path="/login" element={<LoginPage />} />
        <Route path="/verify" element={<VerifyOtpPage />} />

        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </BrowserRouter>
  )
}
