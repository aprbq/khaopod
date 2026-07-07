import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom'
import { AdminLayout } from '@/components/AdminLayout'
import { StoreLayout } from '@/components/StoreLayout'
import { AdminRoute } from '@/features/auth/AdminRoute'
import { ProtectedRoute } from '@/features/auth/ProtectedRoute'
import { AdminDashboardPage } from '@/pages/admin/AdminDashboardPage'
import { AdminOrderDetailPage } from '@/pages/admin/AdminOrderDetailPage'
import { AdminOrdersPage } from '@/pages/admin/AdminOrdersPage'
import { AdminProductFormPage } from '@/pages/admin/AdminProductFormPage'
import { AdminProductsPage } from '@/pages/admin/AdminProductsPage'
import { AdminUsersPage } from '@/pages/admin/AdminUsersPage'
import { CartPage } from '@/pages/CartPage'
import { CheckoutPage } from '@/pages/CheckoutPage'
import { HomePage } from '@/pages/HomePage'
import { LoginPage } from '@/pages/LoginPage'
import { OrderDetailPage } from '@/pages/OrderDetailPage'
import { OrdersPage } from '@/pages/OrdersPage'
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
            <Route path="/checkout" element={<CheckoutPage />} />
            <Route path="/orders" element={<OrdersPage />} />
            <Route path="/orders/:orderNumber" element={<OrderDetailPage />} />
          </Route>
        </Route>

        {/* หลังบ้านแอดมิน (layout แยกจากหน้าร้าน — sidebar ซ้าย) */}
        <Route element={<AdminRoute />}>
          <Route element={<AdminLayout />}>
            <Route path="/admin" element={<AdminDashboardPage />} />
            <Route path="/admin/orders" element={<AdminOrdersPage />} />
            <Route path="/admin/orders/:orderNumber" element={<AdminOrderDetailPage />} />
            <Route path="/admin/products" element={<AdminProductsPage />} />
            <Route path="/admin/products/new" element={<AdminProductFormPage />} />
            <Route path="/admin/products/:id" element={<AdminProductFormPage />} />
            <Route path="/admin/users" element={<AdminUsersPage />} />
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
