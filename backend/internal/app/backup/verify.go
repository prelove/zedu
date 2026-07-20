package backup

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// VerifyPackage validates a package and restores it only into a new target.
// It never modifies the active database or any existing destination.
func VerifyPackage(packageDir, targetDir string) error {
	if _, err := os.Stat(targetDir); err == nil {
		return fmt.Errorf("verification target already exists")
	} else if !os.IsNotExist(err) {
		return err
	}
	if err := verifyManifest(packageDir); err != nil {
		return err
	}
	staging := targetDir + ".staging"
	if err := os.RemoveAll(staging); err != nil {
		return err
	}
	if err := copyTree(packageDir, staging); err != nil {
		_ = os.RemoveAll(staging)
		return err
	}
	if err := verifySQLite(filepath.Join(staging, "zedu.db")); err != nil {
		_ = os.RemoveAll(staging)
		return err
	}
	if err := os.Rename(staging, targetDir); err != nil {
		_ = os.RemoveAll(staging)
		return err
	}
	return nil
}

func verifySQLite(path string) error {
	db, err := sql.Open("sqlite", "file:"+filepath.ToSlash(path)+"?mode=ro")
	if err != nil {
		return err
	}
	defer db.Close()
	var value int
	if err := db.QueryRow(`SELECT 1`).Scan(&value); err != nil {
		return err
	}
	if value != 1 {
		return fmt.Errorf("invalid backup database")
	}
	return nil
}
