package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type databaseManifest struct {
	Databases []DatabaseDocument `json:"databases"`
}

// renderDatabaseDocumentation writes the SPA data contract plus searchable
// Markdown fallbacks. The JSON contains schema metadata only, never row data.
func renderDatabaseDocumentation(project *Project, outDir string) error {
	data, err := json.MarshalIndent(databaseManifest{Databases: project.Databases}, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	dataDir := filepath.Join(outDir, "data")
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dataDir, "databases.json"), data, 0o644); err != nil {
		return err
	}
	if len(project.Databases) == 0 {
		return nil
	}

	databaseDir := filepath.Join(outDir, "database")
	if err := os.MkdirAll(databaseDir, 0o755); err != nil {
		return err
	}
	var index strings.Builder
	index.WriteString("# Databases\n\n")
	index.WriteString("Schema documentation generated from explicitly selected database files.\n\n")
	index.WriteString("| Database | Dialect | Tables | Columns | Relationships |\n")
	index.WriteString("|---|---|---:|---:|---:|\n")
	for _, database := range project.Databases {
		fmt.Fprintf(
			&index,
			"| [%s](%s.md) | %s | %d | %d | %d |\n",
			database.Name,
			database.Slug,
			database.Dialect,
			len(database.Tables),
			countDatabaseColumns(database),
			countDatabaseRelationships(database),
		)
		if err := renderDatabaseMarkdown(database, databaseDir); err != nil {
			return err
		}
	}
	return os.WriteFile(filepath.Join(databaseDir, "index.md"), []byte(index.String()), 0o644)
}

func renderDatabaseMarkdown(database DatabaseDocument, databaseDir string) error {
	var output strings.Builder
	fmt.Fprintf(&output, "# %s data dictionary\n\n", database.Name)
	fmt.Fprintf(&output, "- Dialect: `%s`\n", database.Dialect)
	fmt.Fprintf(&output, "- Source: `%s`\n", database.Source)
	fmt.Fprintf(&output, "- Tables: %d\n", len(database.Tables))
	fmt.Fprintf(&output, "- Views: %d\n", len(database.Views))
	fmt.Fprintf(&output, "- Relationships: %d\n\n", countDatabaseRelationships(database))
	if len(database.Relationships) > 0 {
		output.WriteString("## Relationships\n\n")
		output.WriteString("| Origin | From | References | Evidence |\n")
		output.WriteString("|---|---|---|---|\n")
		for _, relationship := range database.Relationships {
			fmt.Fprintf(
				&output,
				"| %s | `%s.%s` | `%s.%s` | %s |\n",
				relationship.Origin,
				relationship.FromTable,
				strings.Join(relationship.FromColumns, ", "),
				relationship.TargetTable,
				strings.Join(relationship.TargetColumns, ", "),
				relationship.Evidence,
			)
		}
		output.WriteString("\n")
	}

	for _, table := range database.Tables {
		fmt.Fprintf(&output, "## %s\n\n", table.Name)
		output.WriteString("| Column | Type | Nullable | Key | Default | Generated |\n")
		output.WriteString("|---|---|---|---|---|---|\n")
		for _, column := range table.Columns {
			key := ""
			if column.PrimaryKey > 0 {
				key = fmt.Sprintf("PK %d", column.PrimaryKey)
			} else if column.Unique {
				key = "Unique"
			}
			defaultValue := ""
			if column.Default != nil {
				defaultValue = *column.Default
			}
			fmt.Fprintf(
				&output,
				"| `%s` | `%s` | %s | %s | `%s` | %s |\n",
				column.Name,
				column.Type,
				yesNo(column.Nullable),
				key,
				strings.ReplaceAll(defaultValue, "|", "\\|"),
				column.Generated,
			)
		}
		output.WriteString("\n")
		if len(table.ForeignKeys) > 0 {
			output.WriteString("### Foreign keys\n\n")
			output.WriteString("| From | References | On update | On delete |\n")
			output.WriteString("|---|---|---|---|\n")
			for _, foreignKey := range table.ForeignKeys {
				fmt.Fprintf(
					&output,
					"| `%s` | `%s.%s` | %s | %s |\n",
					foreignKey.From,
					foreignKey.TargetTable,
					foreignKey.TargetColumn,
					foreignKey.OnUpdate,
					foreignKey.OnDelete,
				)
			}
			output.WriteString("\n")
		}
	}

	if len(database.Views) > 0 {
		output.WriteString("## Views\n\n")
		for _, view := range database.Views {
			fmt.Fprintf(&output, "### %s\n\n", view.Name)
			if view.InspectionError != "" {
				fmt.Fprintf(
					&output,
					"> Column metadata unavailable: `%s`\n\n",
					strings.ReplaceAll(view.InspectionError, "`", "'"),
				)
			}
			fmt.Fprintf(&output, "```sql\n%s\n```\n\n", view.SQL)
		}
	}
	if len(database.Triggers) > 0 {
		output.WriteString("## Triggers\n\n")
		for _, trigger := range database.Triggers {
			fmt.Fprintf(
				&output,
				"### %s\n\nTable: `%s`\n\n```sql\n%s\n```\n\n",
				trigger.Name,
				trigger.Table,
				trigger.SQL,
			)
		}
	}
	return os.WriteFile(filepath.Join(databaseDir, database.Slug+".md"), []byte(output.String()), 0o644)
}

func yesNo(value bool) string {
	if value {
		return "Yes"
	}
	return "No"
}
