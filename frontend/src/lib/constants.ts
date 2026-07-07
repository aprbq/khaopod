// ข้อมูลรับชำระเงินของร้าน (โชว์หน้า order detail) — TODO: แก้เป็นค่าจริงก่อนเปิดใช้งาน
export const SHOP_PAYMENT = {
  promptpayId: '08X-XXX-XXXX',
  bankName: 'กสิกรไทย',
  bankAccountNo: 'XXX-X-XXXXX-X',
  bankAccountName: 'กองบัญชาการข่าวปด',
}

// ค่าส่งเหมาจ่าย — ใช้โชว์สรุปยอดหน้า checkout เท่านั้น ยอดจริงคิดที่ backend (docs §9.1)
export const FLAT_SHIPPING_FEE = 40
