//go:build unix

package page

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"
)

func TestCreateHonoursUmask(t *testing.T) {
	old := syscall.Umask(0o077)
	defer syscall.Umask(old)

	dir := t.TempDir()
	if err := Create(dir, "secret", []byte("x")); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(filepath.Join(PagesDir(dir), "secret.md"))
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Errorf("mode = %o, want 600 under umask 077", got)
	}
}
