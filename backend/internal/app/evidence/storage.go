package evidence

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

type Storage struct {
	dataRoot   string
	uploadsDir string
	tmpDir     string
	rename     func(string, string) error
	remove     func(string) error
	mkdirAll   func(string, os.FileMode) error
}

type stagedFile struct {
	path string
	size int64
	mime string
	ext  string
}

func NewStorage(dataRoot string) *Storage {
	dataRoot = strings.TrimSpace(dataRoot)
	if dataRoot == "" {
		dataRoot = "data"
	}
	uploadsDir := filepath.Join(dataRoot, "uploads")
	return &Storage{
		dataRoot:   dataRoot,
		uploadsDir: uploadsDir,
		tmpDir:     filepath.Join(uploadsDir, ".tmp"),
		rename:     os.Rename,
		remove:     os.Remove,
		mkdirAll:   os.MkdirAll,
	}
}

func (s *Storage) CleanupTemp() error {
	if err := s.mkdirAll(s.tmpDir, 0o755); err != nil {
		return err
	}
	entries, err := os.ReadDir(s.tmpDir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if err := s.remove(filepath.Join(s.tmpDir, entry.Name())); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}

func (s *Storage) Stage(r io.Reader) (stagedFile, error) {
	if err := s.mkdirAll(s.tmpDir, 0o755); err != nil {
		return stagedFile{}, err
	}

	tmp, err := os.CreateTemp(s.tmpDir, "attachment-*")
	if err != nil {
		return stagedFile{}, err
	}
	tmpPath := tmp.Name()
	cleanup := func() {
		_ = tmp.Close()
		_ = s.remove(tmpPath)
	}

	limited := &io.LimitedReader{R: r, N: maxAttachmentSize + 1}
	written, err := io.Copy(tmp, limited)
	if err != nil {
		cleanup()
		return stagedFile{}, err
	}
	if written == 0 || written > maxAttachmentSize {
		cleanup()
		return stagedFile{}, ErrInvalidState
	}
	if _, err := tmp.Seek(0, io.SeekStart); err != nil {
		cleanup()
		return stagedFile{}, err
	}

	header := make([]byte, 32)
	n, err := tmp.Read(header)
	if err != nil && err != io.EOF {
		cleanup()
		return stagedFile{}, err
	}
	mimeType, ext, ok := detectAllowedFileType(header[:n])
	if !ok {
		cleanup()
		return stagedFile{}, ErrInvalidState
	}
	if err := tmp.Close(); err != nil {
		_ = s.remove(tmpPath)
		return stagedFile{}, err
	}

	return stagedFile{path: tmpPath, size: written, mime: mimeType, ext: ext}, nil
}

func (s *Storage) Publish(staged stagedFile, paymentID, attachmentID int64) (string, error) {
	relPath := path.Join("payments", strconv.FormatInt(paymentID, 10), fmt.Sprintf("%d%s", attachmentID, staged.ext))
	absPath, err := s.resolve(relPath)
	if err != nil {
		return "", err
	}
	if err := s.mkdirAll(filepath.Dir(absPath), 0o755); err != nil {
		return "", err
	}
	if err := s.rename(staged.path, absPath); err != nil {
		return "", err
	}
	return relPath, nil
}

func (s *Storage) DiscardTemp(path string) {
	if path == "" {
		return
	}
	_ = s.remove(path)
}

func (s *Storage) ResolveContentPath(relPath string) (string, error) {
	return s.resolve(relPath)
}

func (s *Storage) resolve(relPath string) (string, error) {
	if relPath == "" || filepath.IsAbs(relPath) {
		return "", fmt.Errorf("invalid attachment path")
	}

	clean := filepath.Clean(filepath.FromSlash(relPath))
	if clean == "." || clean == string(filepath.Separator) {
		return "", fmt.Errorf("invalid attachment path")
	}
	if strings.HasPrefix(clean, ".."+string(filepath.Separator)) || clean == ".." {
		return "", fmt.Errorf("invalid attachment path")
	}

	rootAbs, err := filepath.Abs(s.uploadsDir)
	if err != nil {
		return "", err
	}
	targetAbs, err := filepath.Abs(filepath.Join(rootAbs, clean))
	if err != nil {
		return "", err
	}
	rel, err := filepath.Rel(rootAbs, targetAbs)
	if err != nil {
		return "", err
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("invalid attachment path")
	}
	return targetAbs, nil
}

func detectAllowedFileType(header []byte) (mimeType, ext string, ok bool) {
	if len(header) >= 8 &&
		header[0] == 0x89 &&
		header[1] == 'P' &&
		header[2] == 'N' &&
		header[3] == 'G' &&
		header[4] == 0x0D &&
		header[5] == 0x0A &&
		header[6] == 0x1A &&
		header[7] == 0x0A {
		return "image/png", ".png", true
	}
	if len(header) >= 3 &&
		header[0] == 0xFF &&
		header[1] == 0xD8 &&
		header[2] == 0xFF {
		return "image/jpeg", ".jpg", true
	}
	if len(header) >= 12 &&
		string(header[0:4]) == "RIFF" &&
		string(header[8:12]) == "WEBP" {
		return "image/webp", ".webp", true
	}
	if len(header) >= 5 &&
		string(header[0:5]) == "%PDF-" {
		return "application/pdf", ".pdf", true
	}
	return "", "", false
}

func sanitizeFileName(originalName, ext string) string {
	replaced := strings.ReplaceAll(originalName, "\\", "/")
	base := path.Base(strings.TrimSpace(replaced))
	if base == "." || base == "/" || base == "" {
		return "attachment" + ext
	}

	stem := strings.TrimSuffix(base, path.Ext(base))
	stem = strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z':
			return r
		case r >= 'A' && r <= 'Z':
			return r
		case r >= '0' && r <= '9':
			return r
		case r == '-', r == '_', r == ' ', r > 127:
			return r
		default:
			return -1
		}
	}, stem)
	stem = strings.Trim(stem, " ._-")
	if stem == "" {
		stem = "attachment"
	}
	return stem + ext
}
