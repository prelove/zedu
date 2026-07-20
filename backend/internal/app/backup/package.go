package backup

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const manifestName = "manifest.json"

// Manifest describes every content file in a portable backup package.
type Manifest struct {
	FormatVersion int            `json:"formatVersion"`
	CreatedAt     time.Time      `json:"createdAt"`
	Files         []ManifestFile `json:"files"`
}

type ManifestFile struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
	Size   int64  `json:"size"`
}

// CreatePackage writes a complete package into a private staging directory,
// verifies it, then atomically publishes it below backupDir. It never copies
// environment variables, secrets, or credentials into the package.
func CreatePackage(db *sql.DB, backupDir, dataRoot string) (string, error) {
	if err := os.MkdirAll(filepath.Join(backupDir, ".tmp"), 0o700); err != nil {
		return "", err
	}
	name, err := newPackageName()
	if err != nil {
		return "", err
	}
	staging := filepath.Join(backupDir, ".tmp", name)
	published := filepath.Join(backupDir, name)
	if err := os.MkdirAll(staging, 0o700); err != nil {
		return "", err
	}
	success := false
	defer func() {
		if !success {
			_ = os.RemoveAll(staging)
			_ = os.RemoveAll(published)
		}
	}()

	if _, err := db.Exec(`VACUUM INTO ?`, filepath.Join(staging, "zedu.db")); err != nil {
		return "", err
	}
	if err := copyUploads(dataRoot, staging); err != nil {
		return "", err
	}
	if err := writeConfigSummary(staging); err != nil {
		return "", err
	}
	if err := writeManifest(staging); err != nil {
		return "", err
	}
	if err := verifyManifest(staging); err != nil {
		return "", err
	}
	if err := os.Rename(staging, published); err != nil {
		return "", err
	}
	success = true
	return name, nil
}

func newPackageName() (string, error) {
	var nonce [6]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		return "", err
	}
	return fmt.Sprintf("zedu-%s-%s", time.Now().UTC().Format("20060102T150405Z"), hex.EncodeToString(nonce[:])), nil
}

func copyUploads(dataRoot, packageDir string) error {
	if dataRoot == "" {
		return nil
	}
	source := filepath.Join(dataRoot, "uploads")
	info, err := os.Stat(source)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("uploads path is not a directory")
	}
	return copyTree(source, filepath.Join(packageDir, "uploads"))
}

func copyTree(source, target string) error {
	return filepath.WalkDir(source, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		destination := filepath.Join(target, rel)
		if entry.IsDir() {
			return os.MkdirAll(destination, 0o700)
		}
		if !entry.Type().IsRegular() {
			return fmt.Errorf("unsupported upload entry %q", rel)
		}
		return copyFile(path, destination)
	})
}

func copyFile(source, target string) error {
	in, err := os.Open(source)
	if err != nil {
		return err
	}
	defer in.Close()
	if err := os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
		return err
	}
	out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(out, in)
	closeErr := out.Close()
	if copyErr != nil {
		return copyErr
	}
	return closeErr
}

func writeConfigSummary(packageDir string) error {
	contents, err := json.MarshalIndent(map[string]any{
		"formatVersion": 1,
		"database":      "sqlite",
		"attachments":   "uploads",
	}, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(packageDir, "config-summary.json"), contents, 0o600)
}

func writeManifest(packageDir string) error {
	files, err := listManifestFiles(packageDir)
	if err != nil {
		return err
	}
	contents, err := json.MarshalIndent(Manifest{FormatVersion: 1, CreatedAt: time.Now().UTC(), Files: files}, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(packageDir, manifestName), contents, 0o600)
}

func listManifestFiles(packageDir string) ([]ManifestFile, error) {
	files := make([]ManifestFile, 0)
	err := filepath.WalkDir(packageDir, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(packageDir, path)
		if err != nil {
			return err
		}
		if filepath.ToSlash(rel) == manifestName {
			return nil
		}
		if !entry.Type().IsRegular() {
			return fmt.Errorf("unsupported package entry %q", rel)
		}
		hash, size, err := hashFile(path)
		if err != nil {
			return err
		}
		files = append(files, ManifestFile{Path: filepath.ToSlash(rel), SHA256: hash, Size: size})
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })
	return files, nil
}

func hashFile(path string) (string, int64, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", 0, err
	}
	defer file.Close()
	hash := sha256.New()
	size, err := io.Copy(hash, file)
	if err != nil {
		return "", 0, err
	}
	return hex.EncodeToString(hash.Sum(nil)), size, nil
}

func verifyManifest(packageDir string) error {
	contents, err := os.ReadFile(filepath.Join(packageDir, manifestName))
	if err != nil {
		return err
	}
	var manifest Manifest
	if err := json.Unmarshal(contents, &manifest); err != nil {
		return err
	}
	if manifest.FormatVersion != 1 || len(manifest.Files) == 0 {
		return fmt.Errorf("invalid backup manifest")
	}
	expected := make(map[string]ManifestFile, len(manifest.Files))
	for _, file := range manifest.Files {
		if !isSafeRelativePath(file.Path) {
			return fmt.Errorf("unsafe manifest path")
		}
		expected[file.Path] = file
	}
	actual, err := listManifestFiles(packageDir)
	if err != nil || len(actual) != len(expected) {
		return fmt.Errorf("manifest verification failed")
	}
	for _, file := range actual {
		expectedFile, ok := expected[file.Path]
		if !ok || file.Size != expectedFile.Size || file.SHA256 != expectedFile.SHA256 {
			return fmt.Errorf("manifest verification failed")
		}
	}
	return nil
}

func isSafeRelativePath(path string) bool {
	return path != "" && !filepath.IsAbs(path) && !strings.Contains(filepath.ToSlash(path), "../")
}
