package workspace

import (
	"path/filepath"
	"testing"

	"github.com/repomind/repomind/internal/ir"
	"github.com/repomind/repomind/internal/parser/apiroute"
	"github.com/repomind/repomind/internal/scanner"
)

func TestDetectWorkspacePackages(t *testing.T) {
	root := filepath.Join("..", "..", "testdata", "fixtures", "monorepo")
	scanResult, err := scanner.Scan(root, scanner.Options{})
	if err != nil {
		t.Fatalf("Scan returned error: %v", err)
	}
	routes, routeErrors := apiroute.Extract(root, scanResult)
	if len(routeErrors) != 0 {
		t.Fatalf("route errors = %v, want none", routeErrors)
	}

	packages, errors := Detect(root, scanResult, nil, routes)
	if len(errors) != 0 {
		t.Fatalf("Detect errors = %v, want none", errors)
	}

	assertPackage(t, packages, "apps/web", "@repomind/web", "node", "Next.js, React", "", 0)
	assertPackage(t, packages, "services/api", "repomind-api", "python", "", "FastAPI", 1)
	assertPackage(t, packages, "services/go", "github.com/repomind/example-go", "go", "", "Gin", 1)
	if findPackage(packages, "apps/web").Parent != "." {
		t.Fatalf("apps/web parent = %q, want root package", findPackage(packages, "apps/web").Parent)
	}
	if !contains(findPackage(packages, "apps/web").Dependencies, "services/api") {
		t.Fatalf("apps/web dependencies = %v, want services/api", findPackage(packages, "apps/web").Dependencies)
	}
}

func assertPackage(t *testing.T, packages []ir.PackageInfo, path string, name string, packageType string, frontend string, backend string, routes int) {
	t.Helper()
	for _, pkg := range packages {
		if pkg.Path != path {
			continue
		}
		if pkg.Name != name {
			t.Fatalf("%s name = %q, want %q", path, pkg.Name, name)
		}
		if pkg.Type != packageType {
			t.Fatalf("%s type = %q, want %q", path, pkg.Type, packageType)
		}
		if pkg.Stack.Frontend != frontend {
			t.Fatalf("%s frontend = %q, want %q", path, pkg.Stack.Frontend, frontend)
		}
		if pkg.Stack.Backend != backend {
			t.Fatalf("%s backend = %q, want %q", path, pkg.Stack.Backend, backend)
		}
		if pkg.Routes != routes {
			t.Fatalf("%s routes = %d, want %d", path, pkg.Routes, routes)
		}
		if pkg.Files == 0 {
			t.Fatalf("%s files = 0, want positive count", path)
		}
		return
	}
	t.Fatalf("missing package %s in %+v", path, packages)
}

func findPackage(packages []ir.PackageInfo, path string) ir.PackageInfo {
	for _, pkg := range packages {
		if pkg.Path == path {
			return pkg
		}
	}
	return ir.PackageInfo{}
}

func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
