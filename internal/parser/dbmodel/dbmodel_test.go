package dbmodel

import (
	"path/filepath"
	"testing"

	"github.com/repomind/repomind/internal/ir"
	"github.com/repomind/repomind/internal/scanner"
)

func TestExtractDatabaseModelsFromFixture(t *testing.T) {
	root := filepath.Join("..", "..", "..", "testdata", "fixtures", "db-repo")
	scanResult, err := scanner.Scan(root, scanner.Options{})
	if err != nil {
		t.Fatalf("Scan returned error: %v", err)
	}

	models, errors := Extract(root, scanResult)
	if len(errors) != 0 {
		t.Fatalf("Extract errors = %v, want none", errors)
	}

	assertModel(t, models, "User", "prisma", "prisma/schema.prisma")
	assertModel(t, models, "Order", "prisma", "prisma/schema.prisma")
	assertModel(t, models, "Wallet", "prisma", "prisma/schema.prisma")
	assertModel(t, models, "Customer", "django", "django_app/models.py")
	assertModel(t, models, "Invoice", "django", "django_app/models.py")
	assertModel(t, models, "Account", "sqlalchemy", "sqlalchemy_app/models.py")
	assertModel(t, models, "Profile", "sqlalchemy", "sqlalchemy_app/models.py")
	assertModel(t, models, "User", "sqlmodel", "sqlmodel_app/models.py")
	assertModel(t, models, "SQLModelOrder", "sqlmodel", "sqlmodel_app/models.py")
	assertModel(t, models, "UserEntity", "typeorm", "typeorm/user.entity.ts")
	assertModel(t, models, "OrderEntity", "typeorm", "typeorm/user.entity.ts")
	assertNoModel(t, models, "Settings")
	assertNoModel(t, models, "Payload")
	assertNoModel(t, models, "UserCreate")

	user := findModel(models, "User", "prisma")
	if !hasField(user.Fields, "id", true, false) {
		t.Fatalf("User fields = %+v, want id primary key", user.Fields)
	}
	if !hasField(user.Fields, "email", false, true) {
		t.Fatalf("User fields = %+v, want email unique", user.Fields)
	}
	if !hasRelation(user.Relations, "orders", "Order", "one-to-many") {
		t.Fatalf("User relations = %+v, want orders -> Order one-to-many", user.Relations)
	}

	customer := findModel(models, "Customer", "django")
	if customer.Table != "customers" {
		t.Fatalf("Customer table = %q, want customers", customer.Table)
	}
	if !hasRelation(findModel(models, "Invoice", "django").Relations, "customer", "Customer", "many-to-one") {
		t.Fatalf("Invoice relations = %+v, want customer -> Customer", findModel(models, "Invoice", "django").Relations)
	}

	if !hasRelation(findModel(models, "Profile", "sqlalchemy").Relations, "account_id", "Account", "many-to-one") {
		t.Fatalf("Profile relations = %+v, want account_id -> Account", findModel(models, "Profile", "sqlalchemy").Relations)
	}
	if !hasRelation(findModel(models, "UserEntity", "typeorm").Relations, "orders", "OrderEntity", "one-to-many") {
		t.Fatalf("UserEntity relations = %+v, want orders -> OrderEntity", findModel(models, "UserEntity", "typeorm").Relations)
	}
	if !hasRelation(findModel(models, "User", "sqlmodel").Relations, "orders", "SQLModelOrder", "one-to-many") {
		t.Fatalf("SQLModel User relations = %+v, want orders -> SQLModelOrder", findModel(models, "User", "sqlmodel").Relations)
	}
	if !hasField(findModel(models, "SQLModelOrder", "sqlmodel").Fields, "user_id", false, false) {
		t.Fatalf("SQLModelOrder fields = %+v, want user_id field", findModel(models, "SQLModelOrder", "sqlmodel").Fields)
	}
}

func TestExtractJavaJPAAndGoGORMModels(t *testing.T) {
	root := filepath.Join("..", "..", "..", "testdata", "fixtures", "multilang-repo")
	scanResult, err := scanner.Scan(root, scanner.Options{})
	if err != nil {
		t.Fatalf("Scan returned error: %v", err)
	}

	models, errors := Extract(root, scanResult)
	if len(errors) != 0 {
		t.Fatalf("Extract errors = %v, want none", errors)
	}

	assertModel(t, models, "User", "jpa", "src/main/java/com/example/User.java")
	assertModel(t, models, "User", "gorm", "internal/models/user.go")
	assertModel(t, models, "Order", "gorm", "internal/models/user.go")
	if !hasRelation(findModel(models, "User", "gorm").Relations, "Orders", "Order", "one-to-many") {
		t.Fatalf("GORM User relations = %+v, want Orders -> Order", findModel(models, "User", "gorm").Relations)
	}
}

func TestParseGoGORMWithASTEmbeddedModelAndSelectors(t *testing.T) {
	content := `package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Email string ` + "`gorm:\"uniqueIndex\"`" + `
	Orders []Order
	Profile *profiles.Profile
}

type Payload struct {
	Name string
}
`
	models := parseGoGORM("internal/entities/user.go", content)

	assertModel(t, models, "User", "gorm", "internal/entities/user.go")
	assertNoModel(t, models, "Payload")
	user := findModel(models, "User", "gorm")
	if !hasField(user.Fields, "Email", false, true) {
		t.Fatalf("User fields = %+v, want unique Email", user.Fields)
	}
	if !hasRelation(user.Relations, "Orders", "Order", "one-to-many") {
		t.Fatalf("User relations = %+v, want Orders -> Order", user.Relations)
	}
	if !hasRelation(user.Relations, "Profile", "Profile", "many-to-one") {
		t.Fatalf("User relations = %+v, want Profile -> Profile", user.Relations)
	}
}

func TestParseGoGORMDoesNotTreatPlainDTOAsModel(t *testing.T) {
	content := `package api

type Payload struct {
	ID uint
	Name string
}
`
	models := parseGoGORM("internal/api/payload.go", content)
	if len(models) != 0 {
		t.Fatalf("models = %+v, want none", models)
	}
}

func assertModel(t *testing.T, models []ir.DBModel, name string, source string, file string) {
	t.Helper()
	model := findModel(models, name, source)
	if model.Name == "" {
		t.Fatalf("missing model %s from %s", name, source)
	}
	if model.File != file {
		t.Fatalf("%s file = %q, want %q", name, model.File, file)
	}
	if model.Line <= 0 {
		t.Fatalf("%s line = %d, want positive line", name, model.Line)
	}
	if model.Confidence == "" {
		t.Fatalf("%s confidence is empty", name)
	}
	if model.Evidence == "" {
		t.Fatalf("%s evidence is empty", name)
	}
}

func findModel(models []ir.DBModel, name string, source string) ir.DBModel {
	for _, model := range models {
		if model.Name == name && model.Source == source {
			return model
		}
	}
	return ir.DBModel{}
}

func assertNoModel(t *testing.T, models []ir.DBModel, name string) {
	t.Helper()
	for _, model := range models {
		if model.Name == name {
			t.Fatalf("unexpected model %s from %s in %s", model.Name, model.Source, model.File)
		}
	}
}

func hasField(fields []ir.DBField, name string, primaryKey bool, unique bool) bool {
	for _, field := range fields {
		if field.Name == name && field.PrimaryKey == primaryKey && field.Unique == unique {
			return true
		}
	}
	return false
}

func hasRelation(relations []ir.DBRelation, name string, target string, relationType string) bool {
	for _, relation := range relations {
		if relation.Name == name && relation.Target == target && relation.Type == relationType {
			return true
		}
	}
	return false
}
