package main

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"
)

func TestInspectSQLiteDatabaseCapturesSchemaMetadata(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "catalog.sqlite")
	database, err := sql.Open("sqlite", databasePath)
	if err != nil {
		t.Fatal(err)
	}
	schema := `
		PRAGMA user_version = 7;
		CREATE TABLE teams (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE
		);
		CREATE TABLE people (
			id INTEGER PRIMARY KEY,
			team_id INTEGER,
			first_name TEXT NOT NULL,
			last_name TEXT NOT NULL,
			display_name TEXT GENERATED ALWAYS AS (first_name || ' ' || last_name) STORED,
			FOREIGN KEY (team_id) REFERENCES teams(id) ON UPDATE CASCADE ON DELETE SET NULL
		);
		CREATE INDEX idx_people_name ON people(last_name DESC, first_name);
		CREATE INDEX idx_people_team_active ON people(team_id) WHERE team_id IS NOT NULL;
		CREATE VIEW team_people AS
			SELECT teams.name AS team_name, people.display_name
			FROM people LEFT JOIN teams ON teams.id = people.team_id;
		CREATE VIEW broken_people AS SELECT absent_column FROM people;
		CREATE TRIGGER people_name_required
			BEFORE INSERT ON people
			WHEN NEW.first_name = ''
			BEGIN
				SELECT RAISE(ABORT, 'first name required');
			END;
	`
	if _, err := database.Exec(schema); err != nil {
		database.Close()
		t.Fatal(err)
	}
	if err := database.Close(); err != nil {
		t.Fatal(err)
	}

	document, err := inspectSQLiteDatabase(databasePath, "data/catalog.sqlite")
	if err != nil {
		t.Fatal(err)
	}
	if document.Name != "catalog" || document.Slug != "catalog" {
		t.Fatalf("unexpected identity: %#v", document)
	}
	if document.UserVersion != 7 {
		t.Fatalf("user version = %d, want 7", document.UserVersion)
	}
	if len(document.Tables) != 2 {
		t.Fatalf("tables = %d, want 2", len(document.Tables))
	}
	if len(document.Views) != 2 {
		t.Fatalf("unexpected views: %#v", document.Views)
	}
	brokenView := findDatabaseView(document, "broken_people")
	if brokenView == nil || brokenView.InspectionError == "" || len(brokenView.Columns) != 0 {
		t.Fatalf("broken view metadata was not preserved: %#v", brokenView)
	}
	teamView := findDatabaseView(document, "team_people")
	if teamView == nil || teamView.InspectionError != "" || len(teamView.Columns) != 2 {
		t.Fatalf("valid view metadata is incomplete: %#v", teamView)
	}
	if len(document.Triggers) != 1 || document.Triggers[0].Table != "people" {
		t.Fatalf("unexpected triggers: %#v", document.Triggers)
	}

	people := findDatabaseTable(document, "people")
	if people == nil {
		t.Fatal("people table not found")
	}
	if len(people.ForeignKeys) != 1 {
		t.Fatalf("foreign keys = %d, want 1", len(people.ForeignKeys))
	}
	foreignKey := people.ForeignKeys[0]
	if foreignKey.From != "team_id" || foreignKey.TargetTable != "teams" ||
		foreignKey.TargetColumn != "id" || foreignKey.OnUpdate != "CASCADE" ||
		foreignKey.OnDelete != "SET NULL" {
		t.Fatalf("unexpected foreign key: %#v", foreignKey)
	}
	generated := findDatabaseColumn(people, "display_name")
	if generated == nil || generated.Generated != "stored" || generated.Hidden {
		t.Fatalf("unexpected generated column: %#v", generated)
	}
	if len(people.Indexes) != 2 {
		t.Fatalf("indexes = %d, want 2", len(people.Indexes))
	}
	if countDatabaseRelationships(document) != 1 {
		t.Fatalf("relationships = %d, want 1", countDatabaseRelationships(document))
	}
	if document.Relationships[0].Origin != "declared" {
		t.Fatalf("unexpected relationship origin: %#v", document.Relationships[0])
	}

	teams := findDatabaseTable(document, "teams")
	id := findDatabaseColumn(teams, "id")
	name := findDatabaseColumn(teams, "name")
	if id == nil || !id.AutoIncrement || id.PrimaryKey != 1 {
		t.Fatalf("unexpected primary key: %#v", id)
	}
	if name == nil || !name.Unique || name.Nullable {
		t.Fatalf("unexpected unique column: %#v", name)
	}
}

func TestBuildDatabaseRelationshipsSuggestsOnlyUnambiguousIDColumns(t *testing.T) {
	tables := []DatabaseTable{
		{
			Name: "notes",
			Columns: []DatabaseColumn{
				{Name: "id", Type: "INTEGER", PrimaryKey: 1},
				{Name: "user_id", Type: "INTEGER"},
				{Name: "category_id", Type: "TEXT"},
			},
		},
		{
			Name: "users",
			Columns: []DatabaseColumn{
				{Name: "id", Type: "INTEGER", PrimaryKey: 1},
			},
		},
		{
			Name: "categories",
			Columns: []DatabaseColumn{
				{Name: "id", Type: "INTEGER", PrimaryKey: 1},
			},
		},
	}
	relationships := buildDatabaseRelationships(tables)
	if len(relationships) != 1 {
		t.Fatalf("relationships = %#v, want one compatible suggestion", relationships)
	}
	relationship := relationships[0]
	if relationship.Origin != "suggested" || relationship.FromTable != "notes" ||
		relationship.FromColumns[0] != "user_id" || relationship.TargetTable != "users" ||
		relationship.TargetColumns[0] != "id" {
		t.Fatalf("unexpected suggestion: %#v", relationship)
	}
}

func TestBuildDatabaseRelationshipsSuggestsLegacyNamedKeys(t *testing.T) {
	tables := []DatabaseTable{
		{
			Name: "Customers",
			Columns: []DatabaseColumn{
				{Name: "ID", Type: "INTEGER", PrimaryKey: 1},
			},
		},
		{
			Name: "Invoices",
			Columns: []DatabaseColumn{
				{Name: "InvoiceNo", Type: "INTEGER", PrimaryKey: 1},
				{Name: "CustomerID", Type: "INTEGER"},
			},
		},
		{
			Name: "Products",
			Columns: []DatabaseColumn{
				{Name: "Code", Type: "TEXT", PrimaryKey: 1},
			},
		},
		{
			Name: "InvoiceItems",
			Columns: []DatabaseColumn{
				{Name: "ID", Type: "INTEGER", PrimaryKey: 1},
				{Name: "InvoiceNo", Type: "INTEGER"},
				{Name: "ProductCode", Type: "TEXT"},
			},
		},
	}

	relationships := buildDatabaseRelationships(tables)
	if len(relationships) != 3 {
		t.Fatalf("relationships = %#v, want three legacy-name suggestions", relationships)
	}
	expected := map[string]string{
		"Invoices.CustomerID":      "Customers.ID",
		"InvoiceItems.InvoiceNo":   "Invoices.InvoiceNo",
		"InvoiceItems.ProductCode": "Products.Code",
	}
	for _, relationship := range relationships {
		source := relationship.FromTable + "." + relationship.FromColumns[0]
		target, exists := expected[source]
		if !exists {
			t.Fatalf("unexpected relationship: %#v", relationship)
		}
		actualTarget := relationship.TargetTable + "." + relationship.TargetColumns[0]
		if relationship.Origin != "suggested" || actualTarget != target {
			t.Fatalf("%s targets %s, want %s: %#v", source, actualTarget, target, relationship)
		}
		delete(expected, source)
	}
	if len(expected) != 0 {
		t.Fatalf("missing relationships: %#v", expected)
	}
}

func TestInspectProjectDatabasesResolvesProjectRelativePath(t *testing.T) {
	projectDir := t.TempDir()
	databasePath := filepath.Join(projectDir, "data", "notes.sqlite")
	if err := os.MkdirAll(filepath.Dir(databasePath), 0o755); err != nil {
		t.Fatal(err)
	}
	database, err := sql.Open("sqlite", databasePath)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := database.Exec("CREATE TABLE notes (id INTEGER PRIMARY KEY, body TEXT NOT NULL)"); err != nil {
		database.Close()
		t.Fatal(err)
	}
	if err := database.Close(); err != nil {
		t.Fatal(err)
	}

	documents, err := inspectProjectDatabases(projectDir, []string{"data/notes.sqlite"})
	if err != nil {
		t.Fatal(err)
	}
	if len(documents) != 1 || documents[0].Source != "data/notes.sqlite" {
		t.Fatalf("unexpected documents: %#v", documents)
	}
}

func TestVerifySQLiteHeaderRejectsOtherFiles(t *testing.T) {
	path := filepath.Join(t.TempDir(), "not-a-database.sqlite")
	if err := os.WriteFile(path, []byte("not sqlite"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := verifySQLiteHeader(path); err == nil {
		t.Fatal("expected invalid SQLite header error")
	}
}

func findDatabaseTable(document DatabaseDocument, name string) *DatabaseTable {
	for index := range document.Tables {
		if document.Tables[index].Name == name {
			return &document.Tables[index]
		}
	}
	return nil
}

func findDatabaseColumn(table *DatabaseTable, name string) *DatabaseColumn {
	if table == nil {
		return nil
	}
	for index := range table.Columns {
		if table.Columns[index].Name == name {
			return &table.Columns[index]
		}
	}
	return nil
}

func findDatabaseView(document DatabaseDocument, name string) *DatabaseView {
	for index := range document.Views {
		if document.Views[index].Name == name {
			return &document.Views[index]
		}
	}
	return nil
}
