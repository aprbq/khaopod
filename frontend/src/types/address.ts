// สะท้อน response ของ /addresses ใน docs/rest_api.md (§7)

export interface Address {
  id: number
  recipient_name: string
  phone: string
  address_line: string
  subdistrict: string
  district: string
  province: string
  postal_code: string
  note?: string
  is_default: boolean
}

export interface AddressInput {
  recipient_name: string
  phone: string
  address_line: string
  subdistrict: string
  district: string
  province: string
  postal_code: string
  note?: string
  is_default?: boolean
}
