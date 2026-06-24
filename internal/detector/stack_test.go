package detector

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/repomind/repomind/internal/ir"
	"github.com/repomind/repomind/internal/scanner"
)

func TestDetectStackFromPackageJSON(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "package.json", `{
  "packageManager": "pnpm@9.0.0",
  "dependencies": {
    "next": "^15.0.0",
    "react": "^19.0.0",
    "express": "^4.0.0",
    "@nestjs/core": "^10.0.0",
    "pg": "^8.0.0",
    "redis": "^4.0.0",
    "bullmq": "^5.0.0"
  }
}`)

	scanResult := mustScan(t, root)
	stack, errors := DetectStack(root, scanResult)
	if len(errors) != 0 {
		t.Fatalf("DetectStack errors = %v, want none", errors)
	}

	if stack.Backend != "NestJS, Express" {
		t.Fatalf("Backend = %q, want %q", stack.Backend, "NestJS, Express")
	}
	if stack.Frontend != "Next.js, React" {
		t.Fatalf("Frontend = %q, want %q", stack.Frontend, "Next.js, React")
	}
	if stack.Database != "Postgres" {
		t.Fatalf("Database = %q, want Postgres", stack.Database)
	}
	if stack.Cache != "Redis" {
		t.Fatalf("Cache = %q, want Redis", stack.Cache)
	}
	if stack.Queue != "BullMQ" {
		t.Fatalf("Queue = %q, want BullMQ", stack.Queue)
	}
	if !slices.Contains(stack.PackageManager, "pnpm") {
		t.Fatalf("PackageManager = %v, want pnpm", stack.PackageManager)
	}
	if !slices.Contains(stack.ConfigFiles, "package.json") {
		t.Fatalf("ConfigFiles = %v, want package.json", stack.ConfigFiles)
	}
}

func TestDetectStackFromPythonAndSettingsFiles(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "requirements.txt", `
Django==5.0
fastapi>=0.110
celery[redis]==5.3
psycopg2-binary==2.9
django-redis==5.4
`)
	writeFile(t, root, "pyproject.toml", `
[tool.poetry]
name = "fixture"

[tool.poetry.dependencies]
python = "^3.12"
mysqlclient = "^2.2"
`)
	writeFile(t, root, "project/settings.py", `
INSTALLED_APPS = [
    "django.contrib.admin",
]
`)

	scanResult := mustScan(t, root)
	stack, errors := DetectStack(root, scanResult)
	if len(errors) != 0 {
		t.Fatalf("DetectStack errors = %v, want none", errors)
	}

	if stack.Backend != "Django, FastAPI" {
		t.Fatalf("Backend = %q, want %q", stack.Backend, "Django, FastAPI")
	}
	if stack.Database != "Postgres, MySQL" {
		t.Fatalf("Database = %q, want %q", stack.Database, "Postgres, MySQL")
	}
	if stack.Cache != "Redis" {
		t.Fatalf("Cache = %q, want Redis", stack.Cache)
	}
	if stack.Queue != "Celery" {
		t.Fatalf("Queue = %q, want Celery", stack.Queue)
	}
	for _, want := range []string{"pip", "poetry"} {
		if !slices.Contains(stack.PackageManager, want) {
			t.Fatalf("PackageManager = %v, want %s", stack.PackageManager, want)
		}
	}
	for _, want := range []string{"project/settings.py", "pyproject.toml", "requirements.txt"} {
		if !slices.Contains(stack.ConfigFiles, want) {
			t.Fatalf("ConfigFiles = %v, want %s", stack.ConfigFiles, want)
		}
	}
}

func TestDetectStackFromDockerCompose(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "docker-compose.yml", `
services:
  db:
    image: postgres:16
  mysql:
    image: mysql:8
  cache:
    image: redis:7
  worker:
    command: celery -A app worker
`)
	writeFile(t, root, ".env.example", "DATABASE_URL=postgres://example\n")

	scanResult := mustScan(t, root)
	stack, errors := DetectStack(root, scanResult)
	if len(errors) != 0 {
		t.Fatalf("DetectStack errors = %v, want none", errors)
	}

	if stack.Database != "Postgres, MySQL" {
		t.Fatalf("Database = %q, want %q", stack.Database, "Postgres, MySQL")
	}
	if stack.Cache != "Redis" {
		t.Fatalf("Cache = %q, want Redis", stack.Cache)
	}
	if stack.Queue != "Celery" {
		t.Fatalf("Queue = %q, want Celery", stack.Queue)
	}
	for _, want := range []string{".env.example", "docker-compose.yml"} {
		if !slices.Contains(stack.ConfigFiles, want) {
			t.Fatalf("ConfigFiles = %v, want %s", stack.ConfigFiles, want)
		}
	}
}

func TestDetectStackFromPHPJavaAndGo(t *testing.T) {
	root := filepath.Join("..", "..", "testdata", "fixtures", "multilang-repo")
	scanResult := mustScan(t, root)

	stack, errors := DetectStack(root, scanResult)
	if len(errors) != 0 {
		t.Fatalf("DetectStack errors = %v, want none", errors)
	}

	for _, want := range []string{"Laravel", "Spring Boot", "Gin"} {
		if !strings.Contains(stack.Backend, want) {
			t.Fatalf("Backend = %q, want %s", stack.Backend, want)
		}
	}
	for _, want := range []string{"Postgres", "MySQL"} {
		if !strings.Contains(stack.Database, want) {
			t.Fatalf("Database = %q, want %s", stack.Database, want)
		}
	}
	if stack.Cache != "Redis" {
		t.Fatalf("Cache = %q, want Redis", stack.Cache)
	}
	for _, want := range []string{"composer", "maven", "go"} {
		if !slices.Contains(stack.PackageManager, want) {
			t.Fatalf("PackageManager = %v, want %s", stack.PackageManager, want)
		}
	}
}

func TestDetectStackFromChiGoMod(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "go.mod", `module github.com/go-chi/chi/v5

go 1.24
`)

	scanResult := mustScan(t, root)
	stack, errors := DetectStack(root, scanResult)
	if len(errors) != 0 {
		t.Fatalf("DetectStack errors = %v, want none", errors)
	}

	if stack.Backend != "Chi" {
		t.Fatalf("Backend = %q, want Chi", stack.Backend)
	}
	if !slices.Contains(stack.PackageManager, "go") {
		t.Fatalf("PackageManager = %v, want go", stack.PackageManager)
	}
}

func mustScan(t *testing.T, root string) ir.ScanSummary {
	t.Helper()

	result, err := scanner.Scan(root, scanner.Options{})
	if err != nil {
		t.Fatalf("Scan returned error: %v", err)
	}
	return result
}

func writeFile(t *testing.T, root string, rel string, content string) {
	t.Helper()

	path := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("create parent directory: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
}
