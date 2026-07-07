// พจนานุกรมข้อความสองภาษา (ไทย/อังกฤษ)
// เพิ่มข้อความใหม่ที่นี่เสมอ อย่า hardcode ในคอมโพเนนต์ — เรียกผ่าน t('key')

export type Lang = 'th' | 'en'

const th = {
  // navbar
  'nav.home': 'หน้าแรก',
  'nav.collection': 'คอลเลกชัน',
  'nav.stickers': 'สติ๊กเกอร์',
  'nav.brandHome': 'กองบัญชาการข่าวปด หน้าแรก',
  'nav.openMenu': 'เปิดเมนู',
  'nav.account': 'บัญชีของฉัน',
  'nav.cart': 'ตะกร้าสินค้า',

  // common
  'common.viewAll': 'ดูทั้งหมด',
  'common.items': '{n} รายการ',
  'common.close': 'ปิด',
  'common.cancel': 'ยกเลิก',
  'common.error': 'เกิดข้อผิดพลาด กรุณาลองใหม่',

  // footer
  'footer.tagline': 'ร้านค้าอย่างเป็นทางการของเพจกองบัญชาการข่าวปด',
  'footer.shop': 'ช้อป',
  'footer.help': 'ช่วยเหลือ',
  'footer.follow': 'ติดตาม',
  'footer.allCollection': 'คอลเลกชั่นทั้งหมด',
  'footer.new': 'มาใหม่',
  'footer.bestSeller': 'ขายดี',
  'footer.myAccount': 'บัญชีของฉัน',
  'footer.howToOrder': 'วิธีสั่งซื้อ',
  'footer.shipping': 'การจัดส่ง',
  'footer.rights': 'สงวนลิขสิทธิ์.',

  // home
  'home.shopNow': 'ช้อปเลย',
  'home.new': 'มาใหม่',

  // shop
  'shop.collection': 'คอลเลกชัน',

  // product grid
  'grid.loadError': 'โหลดสินค้าไม่สำเร็จ กรุณาลองใหม่อีกครั้ง',
  'grid.empty': 'ยังไม่มีสินค้า',

  // product detail
  'product.notFound': 'ไม่พบสินค้านี้',
  'product.backCollection': 'กลับไปหน้าคอลเลกชั่น',
  'product.backCollectionShort': '← คอลเลกชั่นทั้งหมด',
  'product.options': 'ตัวเลือก',
  'product.size': 'ไซซ์',
  'product.color': 'สี',
  'product.outOfStock': 'สินค้าหมด',
  'product.selectOption': 'เลือกตัวเลือกก่อน',
  'product.addToCart': 'เพิ่มลงตะกร้า',
  'product.cartComingSoon': 'ระบบตะกร้ากำลังจะมาเร็ว ๆ นี้ — ตอนนี้ยังเป็นเดโมหน้าร้าน',
  'product.viewLarge': 'ดูรูปขนาดใหญ่',
  'product.viewImageN': 'ดูรูปที่ {n}',
  'product.added': 'เพิ่มลงตะกร้าแล้ว',
  'product.viewCart': 'ดูตะกร้า',

  // cart
  'cart.emptyTitle': 'ตะกร้าของคุณยังว่าง',
  'cart.emptyDesc': 'ยังไม่มีสินค้าในตะกร้า เลือกซื้อได้เลย',
  'cart.browse': 'เลือกซื้อสินค้า',
  'cart.docTitle': 'ตะกร้า',
  'cart.title': 'ตะกร้าสินค้า',
  'cart.subtotal': 'ยอดรวม',
  'cart.checkout': 'ดำเนินการสั่งซื้อ',
  'cart.checkoutSoon': 'ระบบชำระเงินกำลังจะมาเร็ว ๆ นี้',
  'cart.remove': 'ลบ',
  'cart.removeConfirmTitle': 'ลบสินค้าออกจากตะกร้า?',
  'cart.removeConfirmDesc': 'ต้องการลบ "{name}" ออกจากตะกร้าหรือไม่',
  'cart.clear': 'ล้างตะกร้า',
  'cart.clearConfirmTitle': 'ล้างตะกร้าทั้งหมด?',
  'cart.clearConfirmDesc': 'สินค้าทุกชิ้นจะถูกลบออกจากตะกร้า',
  'cart.outOfStock': 'สต็อกไม่พอ',
  'cart.loginRequired': 'กรุณาเข้าสู่ระบบเพื่อดูตะกร้า',
  'cart.login': 'เข้าสู่ระบบ',

  // login
  'login.brand': 'กองบัญชาการข่าวปด',
  'login.title': 'เข้าสู่ระบบ',
  'login.desc': 'กรอกอีเมลเพื่อรับรหัส OTP — ไม่ต้องใช้รหัสผ่าน',
  'login.email': 'อีเมล',
  'login.emailInvalid': 'อีเมลไม่ถูกต้อง',
  'login.requestOtp': 'ขอรหัส OTP',

  // verify
  'verify.title': 'ยืนยันรหัส OTP',
  'verify.docTitle': 'ยืนยัน OTP',
  'verify.sentTo': 'ส่งรหัส 6 หลักไปที่ {email} แล้ว',
  'verify.otp': 'รหัส OTP',
  'verify.otpInvalid': 'รหัส OTP ต้องเป็นตัวเลข 6 หลัก',
  'verify.expiresIn': 'รหัสหมดอายุใน {time} นาที',
  'verify.expired': 'รหัสหมดอายุแล้ว กรุณาขอรหัสใหม่',
  'verify.submit': 'ยืนยันและเข้าสู่ระบบ',
  'verify.resendIn': 'ขอรหัสใหม่ได้ใน {sec} วิ',
  'verify.resend': 'ขอรหัสใหม่',

  // profile
  'profile.title': 'บัญชีของฉัน',
  'profile.docTitle': 'โปรไฟล์',
  'profile.logout': 'ออกจากระบบ',
  'profile.roleAdmin': 'ผู้ดูแลระบบ',
  'profile.joined': 'เข้าร่วมเมื่อ {date}',
  'profile.lastUpdated': 'แก้ไขข้อมูลล่าสุด {date}',
  'profile.changeAvatar': 'เปลี่ยนรูปโปรไฟล์',
  'profile.avatarTooBig': 'ไฟล์รูปต้องไม่เกิน 2MB',
  'profile.editProfile': 'แก้ไขโปรไฟล์',
  'profile.displayName': 'ชื่อที่แสดง',
  'profile.displayNamePh': 'เช่น สมชาย ใจดี',
  'profile.nameTooLong': 'ชื่อยาวเกินไป',
  'profile.phone': 'เบอร์โทร',
  'profile.phonePh': '0812345678',
  'profile.phoneInvalid': 'เบอร์โทรต้องขึ้นต้น 0 และมี 10 หลัก',
  'profile.saved': 'บันทึกโปรไฟล์แล้ว',
  'profile.save': 'บันทึก',
}

export type TranslationKey = keyof typeof th

const en: Record<TranslationKey, string> = {
  'nav.home': 'Home',
  'nav.collection': 'Collection',
  'nav.stickers': 'Stickers',
  'nav.brandHome': 'Khaopod News HQ home',
  'nav.openMenu': 'Open menu',
  'nav.account': 'My account',
  'nav.cart': 'Shopping cart',

  'common.viewAll': 'View all',
  'common.items': '{n} items',
  'common.close': 'Close',
  'common.cancel': 'Cancel',
  'common.error': 'Something went wrong, please try again',

  'footer.tagline': 'The official store of the Khaopod News page',
  'footer.shop': 'Shop',
  'footer.help': 'Help',
  'footer.follow': 'Follow',
  'footer.allCollection': 'All collection',
  'footer.new': 'New arrivals',
  'footer.bestSeller': 'Best sellers',
  'footer.myAccount': 'My account',
  'footer.howToOrder': 'How to order',
  'footer.shipping': 'Shipping',
  'footer.rights': 'All rights reserved.',

  'home.shopNow': 'Shop now',
  'home.new': 'New arrivals',

  'shop.collection': 'Collection',

  'grid.loadError': 'Failed to load products. Please try again.',
  'grid.empty': 'No products yet',

  'product.notFound': 'Product not found',
  'product.backCollection': 'Back to collection',
  'product.backCollectionShort': '← All collection',
  'product.options': 'Options',
  'product.size': 'Size',
  'product.color': 'Color',
  'product.outOfStock': 'Out of stock',
  'product.selectOption': 'Select an option',
  'product.addToCart': 'Add to cart',
  'product.cartComingSoon': 'Cart is coming soon — this is a store demo for now',
  'product.viewLarge': 'View larger image',
  'product.viewImageN': 'View image {n}',
  'product.added': 'Added to cart',
  'product.viewCart': 'View cart',

  'cart.emptyTitle': 'Your cart is empty',
  'cart.emptyDesc': 'No items in your cart yet — start shopping',
  'cart.browse': 'Browse products',
  'cart.docTitle': 'Cart',
  'cart.title': 'Shopping cart',
  'cart.subtotal': 'Subtotal',
  'cart.checkout': 'Checkout',
  'cart.checkoutSoon': 'Checkout is coming soon',
  'cart.remove': 'Remove',
  'cart.removeConfirmTitle': 'Remove item from cart?',
  'cart.removeConfirmDesc': 'Remove "{name}" from your cart?',
  'cart.clear': 'Clear cart',
  'cart.clearConfirmTitle': 'Clear entire cart?',
  'cart.clearConfirmDesc': 'All items will be removed from your cart',
  'cart.outOfStock': 'Not enough stock',
  'cart.loginRequired': 'Please sign in to view your cart',
  'cart.login': 'Sign in',

  'login.brand': 'Khaopod News',
  'login.title': 'Sign in',
  'login.desc': 'Enter your email to get an OTP — no password needed',
  'login.email': 'Email',
  'login.emailInvalid': 'Invalid email',
  'login.requestOtp': 'Request OTP',

  'verify.title': 'Verify OTP',
  'verify.docTitle': 'Verify OTP',
  'verify.sentTo': 'A 6-digit code was sent to {email}',
  'verify.otp': 'OTP code',
  'verify.otpInvalid': 'OTP must be 6 digits',
  'verify.expiresIn': 'Code expires in {time}',
  'verify.expired': 'Code expired, please request a new one',
  'verify.submit': 'Verify & sign in',
  'verify.resendIn': 'Resend in {sec}s',
  'verify.resend': 'Resend code',

  'profile.title': 'My account',
  'profile.docTitle': 'Profile',
  'profile.logout': 'Sign out',
  'profile.roleAdmin': 'Admin',
  'profile.joined': 'Joined {date}',
  'profile.lastUpdated': 'Last updated {date}',
  'profile.changeAvatar': 'Change photo',
  'profile.avatarTooBig': 'Image must be under 2MB',
  'profile.editProfile': 'Edit profile',
  'profile.displayName': 'Display name',
  'profile.displayNamePh': 'e.g. John Doe',
  'profile.nameTooLong': 'Name is too long',
  'profile.phone': 'Phone',
  'profile.phonePh': '0812345678',
  'profile.phoneInvalid': 'Phone must start with 0 and be 10 digits',
  'profile.saved': 'Profile saved',
  'profile.save': 'Save',
}

export const translations: Record<Lang, Record<TranslationKey, string>> = { th, en }
