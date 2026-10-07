package catalog

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveRejectsPaths(t *testing.T) {
	catalog := &Catalog{Root: t.TempDir()}
	for _, name := range []string{"", ".", "..", "../outside", "nested/lab"} {
		t.Run(name, func(t *testing.T) {
			if _, err := catalog.Resolve(name); err == nil {
				t.Fatalf("expected %q to be rejected", name)
			}
		})
	}
}

func TestFindWalksToRepositoryRoot(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "labs"), 0o755); err != nil {
		t.Fatal(err)
	}
	child := filepath.Join(root, "nested", "directory")
	if err := os.MkdirAll(child, 0o755); err != nil {
		t.Fatal(err)
	}

	catalog, err := Find(func() (string, error) {
		return child, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if catalog.Root != filepath.Join(root, "labs") {
		t.Fatalf("catalog root = %q", catalog.Root)
	}
}
