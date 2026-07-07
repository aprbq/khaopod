// Package storage — outbound adapter เก็บไฟล์อัปโหลดลงดิสก์ (implements output.FileStorage)
// ไฟล์ถูกเสิร์ฟกลับผ่าน static route ของ router (publicPrefix เช่น /uploads)
package storage

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/khaopod/backend/internal/core/port/output"
)

type LocalStorage struct {
	baseDir      string
	publicPrefix string // prefix ของ URL ที่ router เสิร์ฟ baseDir (เช่น "/uploads")
}

var _ output.FileStorage = (*LocalStorage)(nil)

func NewLocal(baseDir, publicPrefix string) (*LocalStorage, error) {
	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		return nil, fmt.Errorf("create upload dir: %w", err)
	}
	return &LocalStorage{baseDir: baseDir, publicPrefix: strings.TrimSuffix(publicPrefix, "/")}, nil
}

func (s *LocalStorage) Save(_ context.Context, relPath string, content []byte) (string, error) {
	clean, err := s.safeRel(relPath)
	if err != nil {
		return "", err
	}
	full := filepath.Join(s.baseDir, clean)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		return "", err
	}
	if err := os.WriteFile(full, content, 0o644); err != nil {
		return "", err
	}
	return s.publicPrefix + "/" + filepath.ToSlash(clean), nil
}

func (s *LocalStorage) Remove(_ context.Context, url string) error {
	rel, ok := strings.CutPrefix(url, s.publicPrefix+"/")
	if !ok {
		return nil // URL ภายนอก (เช่นรูปโปรไฟล์จาก Google) — ไม่ใช่ไฟล์ของเรา
	}
	clean, err := s.safeRel(rel)
	if err != nil {
		return nil // URL เพี้ยน — ไม่มีไฟล์ให้ลบ
	}
	if err := os.Remove(filepath.Join(s.baseDir, clean)); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

// safeRel กัน path traversal — รับเฉพาะ relative path ที่ไม่หนีออกนอก baseDir
func (s *LocalStorage) safeRel(p string) (string, error) {
	clean := filepath.Clean(filepath.FromSlash(p))
	if filepath.IsAbs(clean) || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("invalid storage path: %q", p)
	}
	return clean, nil
}
