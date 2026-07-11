package httpserver_test

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const modulePath = "github.com/prelove/zedu/backend"

func TestDependencyBoundaries(t *testing.T) {
	root := filepath.Join("..", "..", "..")
	root, err := filepath.Abs(root)
	if err != nil {
		t.Fatalf("resolve backend root: %v", err)
	}

	checkNoImports(t, filepath.Join(root, "pkg"), modulePath+"/internal/")
	checkPlatformImports(t, filepath.Join(root, "internal", "platform"))
}

func checkNoImports(t *testing.T, dir string, forbidden string) {
	t.Helper()
	_ = walkGoFiles(dir, func(path string) error {
		imports, err := extractImports(path)
		if err != nil {
			return err
		}
		for _, imp := range imports {
			if strings.HasPrefix(imp, forbidden) {
				t.Errorf("pkg package imports forbidden internal path: %s imports %q", path, imp)
			}
		}
		return nil
	})
}

func checkPlatformImports(t *testing.T, dir string) {
	t.Helper()
	allowedInternal := modulePath + "/internal/platform/"
	_ = walkGoFiles(dir, func(path string) error {
		imports, err := extractImports(path)
		if err != nil {
			return err
		}
		for _, imp := range imports {
			if strings.HasPrefix(imp, modulePath+"/internal/") && !strings.HasPrefix(imp, allowedInternal) {
				t.Errorf("platform package imports business/internal module: %s imports %q", path, imp)
			}
		}
		return nil
	})
}

func extractImports(path string) ([]string, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
	if err != nil {
		return nil, err
	}
	var imports []string
	for _, imp := range f.Imports {
		v := strings.Trim(imp.Path.Value, `"`)
		imports = append(imports, v)
	}
	return imports, nil
}

func walkGoFiles(dir string, fn func(string) error) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		return fn(path)
	})
}
