// zedu-backup-verify verifies a portable backup and restores it only into a
// caller-supplied new directory. It never opens or replaces the active DB.
package main

import (
	"fmt"
	"os"

	"github.com/prelove/zedu/backend/internal/app/backup"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: zedu-backup-verify <package-dir> <new-target-dir>")
		os.Exit(2)
	}
	if err := backup.VerifyPackage(os.Args[1], os.Args[2]); err != nil {
		fmt.Fprintln(os.Stderr, "backup verification failed")
		os.Exit(1)
	}
	fmt.Println("backup verification passed")
}
