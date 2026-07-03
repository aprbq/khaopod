-- ย้อน migration 0001
DROP TABLE IF EXISTS auth_sessions;
DROP TABLE IF EXISTS otp_codes;
DROP TABLE IF EXISTS user_oauth_accounts;
DROP TABLE IF EXISTS users;
DROP TYPE IF EXISTS otp_purpose;
-- ไม่ DROP FUNCTION set_updated_at / extensions เพราะตารางอื่นอาจใช้ร่วม
