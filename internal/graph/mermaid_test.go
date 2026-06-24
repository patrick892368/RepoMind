package graph

import (
	"strings"
	"testing"

	"github.com/repomind/repomind/internal/ir"
)

func TestGenerateER(t *testing.T) {
	diagram := GenerateER([]ir.DBModel{
		{
			Name:   "User",
			Fields: []ir.DBField{{Name: "id", Type: "Int", PrimaryKey: true}},
			Relations: []ir.DBRelation{
				{Name: "orders", Target: "Order", Type: "one-to-many"},
			},
		},
		{
			Name:   "Order",
			Fields: []ir.DBField{{Name: "id", Type: "Int", PrimaryKey: true}},
		},
	})

	for _, want := range []string{
		"erDiagram",
		"User {",
		"Int id PK",
		"User ||--o{ Order : orders",
	} {
		if !strings.Contains(diagram, want) {
			t.Fatalf("diagram did not contain %q:\n%s", want, diagram)
		}
	}
}

func TestGenerateAPI(t *testing.T) {
	diagram := GenerateAPI([]ir.APIRoute{
		{
			Method:  "POST",
			Path:    "/login",
			Handler: "login",
			File:    "app/main.py",
			Source:  "fastapi",
		},
	})

	for _, want := range []string{
		"flowchart LR",
		"Client",
		"POST /login",
		"app/main.py:login",
	} {
		if !strings.Contains(diagram, want) {
			t.Fatalf("diagram did not contain %q:\n%s", want, diagram)
		}
	}
}

func TestGenerateCallGraph(t *testing.T) {
	diagram := GenerateCallGraph([]ir.CallEdge{
		{Caller: "pay_callback", Callee: "update_order", File: "payment/flow.py"},
	})

	for _, want := range []string{"flowchart TD", "pay_callback", "update_order"} {
		if !strings.Contains(diagram, want) {
			t.Fatalf("diagram did not contain %q:\n%s", want, diagram)
		}
	}
}

func TestGeneratePackageGraph(t *testing.T) {
	diagram := GeneratePackageGraph([]ir.PackageInfo{
		{Path: ".", Name: "root"},
		{Path: "apps/web", Name: "web", Parent: ".", Dependencies: []string{"services/api"}},
		{Path: "services/api", Name: "api", Parent: "."},
	})

	for _, want := range []string{"flowchart TD", "pkg_root", "pkg_apps_web", "pkg_services_api", "pkg_root --> pkg_apps_web", "pkg_apps_web -.-> pkg_services_api"} {
		if !strings.Contains(diagram, want) {
			t.Fatalf("diagram did not contain %q:\n%s", want, diagram)
		}
	}
}
