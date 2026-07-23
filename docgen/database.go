package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"

	_ "modernc.org/sqlite"
)

// DatabaseDocument is the portable schema model emitted for the editorial
// data dictionary and ER viewer. It intentionally contains schema metadata,
// never application row data.
type DatabaseDocument struct {
	Name          string                 `json:"name"`
	Slug          string                 `json:"slug"`
	Dialect       string                 `json:"dialect"`
	Source        string                 `json:"source"`
	FileSize      int64                  `json:"fileSize"`
	UserVersion   int                    `json:"userVersion"`
	Tables        []DatabaseTable        `json:"tables"`
	Views         []DatabaseView         `json:"views"`
	Triggers      []DatabaseTrigger      `json:"triggers"`
	Relationships []DatabaseRelationship `json:"relationships"`
}

type DatabaseTable struct {
	Name         string               `json:"name"`
	SQL          string               `json:"sql,omitempty"`
	Strict       bool                 `json:"strict"`
	WithoutRowID bool                 `json:"withoutRowId"`
	Virtual      bool                 `json:"virtual"`
	Columns      []DatabaseColumn     `json:"columns"`
	Indexes      []DatabaseIndex      `json:"indexes"`
	ForeignKeys  []DatabaseForeignKey `json:"foreignKeys"`
}

type DatabaseColumn struct {
	Position      int     `json:"position"`
	Name          string  `json:"name"`
	Type          string  `json:"type"`
	Nullable      bool    `json:"nullable"`
	Default       *string `json:"default,omitempty"`
	PrimaryKey    int     `json:"primaryKey"`
	Unique        bool    `json:"unique"`
	AutoIncrement bool    `json:"autoIncrement"`
	Hidden        bool    `json:"hidden"`
	Generated     string  `json:"generated,omitempty"`
}

type DatabaseIndex struct {
	Name    string                `json:"name"`
	Unique  bool                  `json:"unique"`
	Origin  string                `json:"origin"`
	Partial bool                  `json:"partial"`
	SQL     string                `json:"sql,omitempty"`
	Columns []DatabaseIndexColumn `json:"columns"`
}

type DatabaseIndexColumn struct {
	Position   int    `json:"position"`
	TableRank  int    `json:"tableRank"`
	Name       string `json:"name,omitempty"`
	Descending bool   `json:"descending"`
	Collation  string `json:"collation,omitempty"`
	Key        bool   `json:"key"`
}

type DatabaseForeignKey struct {
	ID           int    `json:"id"`
	Sequence     int    `json:"sequence"`
	From         string `json:"from"`
	TargetTable  string `json:"targetTable"`
	TargetColumn string `json:"targetColumn"`
	OnUpdate     string `json:"onUpdate"`
	OnDelete     string `json:"onDelete"`
	Match        string `json:"match"`
}

type DatabaseView struct {
	Name            string           `json:"name"`
	SQL             string           `json:"sql,omitempty"`
	Columns         []DatabaseColumn `json:"columns"`
	InspectionError string           `json:"inspectionError,omitempty"`
}

type DatabaseTrigger struct {
	Name  string `json:"name"`
	Table string `json:"table"`
	SQL   string `json:"sql,omitempty"`
}

type DatabaseRelationship struct {
	ID            string   `json:"id"`
	FromTable     string   `json:"fromTable"`
	FromColumns   []string `json:"fromColumns"`
	TargetTable   string   `json:"targetTable"`
	TargetColumns []string `json:"targetColumns"`
	Origin        string   `json:"origin"`
	OnUpdate      string   `json:"onUpdate,omitempty"`
	OnDelete      string   `json:"onDelete,omitempty"`
	Evidence      string   `json:"evidence"`
}

// inspectProjectDatabases resolves explicit paths against the Xojo project
// directory and returns deterministic schema documents.
func inspectProjectDatabases(projectDir string, databasePaths []string) ([]DatabaseDocument, error) {
	if len(databasePaths) == 0 {
		return nil, nil
	}

	seenPaths := map[string]bool{}
	seenSlugs := map[string]bool{}
	documents := make([]DatabaseDocument, 0, len(databasePaths))
	for _, requestedPath := range databasePaths {
		requestedPath = strings.TrimSpace(requestedPath)
		if requestedPath == "" {
			continue
		}
		resolvedPath := filepath.FromSlash(requestedPath)
		if !filepath.IsAbs(resolvedPath) {
			resolvedPath = filepath.Join(projectDir, resolvedPath)
		}
		absolutePath, err := filepath.Abs(resolvedPath)
		if err != nil {
			return nil, fmt.Errorf("database %q: resolve path: %w", requestedPath, err)
		}
		if seenPaths[absolutePath] {
			continue
		}
		seenPaths[absolutePath] = true

		sourceLabel := filepath.Base(absolutePath)
		if relativePath, err := filepath.Rel(projectDir, absolutePath); err == nil &&
			relativePath != ".." && !strings.HasPrefix(relativePath, ".."+string(filepath.Separator)) {
			sourceLabel = filepath.ToSlash(relativePath)
		}
		document, err := inspectSQLiteDatabase(absolutePath, sourceLabel)
		if err != nil {
			return nil, fmt.Errorf("database %q: %w", requestedPath, err)
		}
		if seenSlugs[document.Slug] {
			return nil, fmt.Errorf("database %q: duplicate output slug %q", requestedPath, document.Slug)
		}
		seenSlugs[document.Slug] = true
		documents = append(documents, document)
	}
	sort.Slice(documents, func(first int, second int) bool {
		return documents[first].Name < documents[second].Name
	})
	return documents, nil
}

func inspectSQLiteDatabase(path string, sourceLabel string) (DatabaseDocument, error) {
	info, err := os.Stat(path)
	if err != nil {
		return DatabaseDocument{}, err
	}
	if info.IsDir() {
		return DatabaseDocument{}, fmt.Errorf("path is a directory")
	}
	if err := verifySQLiteHeader(path); err != nil {
		return DatabaseDocument{}, err
	}

	absolutePath, err := filepath.Abs(path)
	if err != nil {
		return DatabaseDocument{}, err
	}
	dsn := (&url.URL{Scheme: "file", Path: absolutePath}).String() + "?mode=ro"
	database, err := sql.Open("sqlite", dsn)
	if err != nil {
		return DatabaseDocument{}, err
	}
	defer database.Close()
	database.SetMaxOpenConns(1)
	if err := database.Ping(); err != nil {
		return DatabaseDocument{}, fmt.Errorf("open read-only: %w", err)
	}
	if _, err := database.Exec("PRAGMA query_only = ON"); err != nil {
		return DatabaseDocument{}, fmt.Errorf("enable query-only mode: %w", err)
	}

	name := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	document := DatabaseDocument{
		Name:          name,
		Slug:          slugify(name),
		Dialect:       "sqlite",
		Source:        filepath.ToSlash(sourceLabel),
		FileSize:      info.Size(),
		Tables:        []DatabaseTable{},
		Views:         []DatabaseView{},
		Triggers:      []DatabaseTrigger{},
		Relationships: []DatabaseRelationship{},
	}
	if err := database.QueryRow("PRAGMA user_version").Scan(&document.UserVersion); err != nil {
		return DatabaseDocument{}, fmt.Errorf("read user_version: %w", err)
	}

	rows, err := database.Query(`
		SELECT name, type, wr, strict
		FROM pragma_table_list
		WHERE schema = 'main'
		  AND name NOT LIKE 'sqlite_%'
		  AND type IN ('table', 'virtual', 'view')
		ORDER BY type, name`)
	if err != nil {
		return DatabaseDocument{}, fmt.Errorf("list tables and views: %w", err)
	}

	type tableListEntry struct {
		name         string
		objectType   string
		withoutRowID bool
		strict       bool
	}
	var entries []tableListEntry
	for rows.Next() {
		var name, objectType string
		var withoutRowID, strict int
		if err := rows.Scan(&name, &objectType, &withoutRowID, &strict); err != nil {
			rows.Close()
			return DatabaseDocument{}, err
		}
		entries = append(entries, tableListEntry{
			name:         name,
			objectType:   objectType,
			withoutRowID: withoutRowID != 0,
			strict:       strict != 0,
		})
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return DatabaseDocument{}, err
	}
	if err := rows.Close(); err != nil {
		return DatabaseDocument{}, err
	}

	for _, entry := range entries {
		name := entry.name
		objectType := entry.objectType
		schemaSQL, err := sqliteSchemaSQL(database, objectTypeForSchema(objectType), name)
		if err != nil {
			return DatabaseDocument{}, err
		}
		if objectType == "view" {
			columns, columnErr := inspectSQLiteColumns(database, name, schemaSQL)
			if columns == nil {
				columns = []DatabaseColumn{}
			}
			inspectionError := ""
			if columnErr != nil {
				inspectionError = columnErr.Error()
			}
			document.Views = append(document.Views, DatabaseView{
				Name:            name,
				SQL:             schemaSQL,
				Columns:         columns,
				InspectionError: inspectionError,
			})
			continue
		}
		columns, err := inspectSQLiteColumns(database, name, schemaSQL)
		if err != nil {
			return DatabaseDocument{}, fmt.Errorf("%s columns: %w", name, err)
		}

		indexes, err := inspectSQLiteIndexes(database, name)
		if err != nil {
			return DatabaseDocument{}, fmt.Errorf("%s indexes: %w", name, err)
		}
		foreignKeys, err := inspectSQLiteForeignKeys(database, name)
		if err != nil {
			return DatabaseDocument{}, fmt.Errorf("%s foreign keys: %w", name, err)
		}
		markUniqueColumns(columns, indexes)
		document.Tables = append(document.Tables, DatabaseTable{
			Name:         name,
			SQL:          schemaSQL,
			Strict:       entry.strict,
			WithoutRowID: entry.withoutRowID,
			Virtual:      objectType == "virtual",
			Columns:      columns,
			Indexes:      indexes,
			ForeignKeys:  foreignKeys,
		})
	}
	triggerRows, err := database.Query(`
		SELECT name, tbl_name, COALESCE(sql, '')
		FROM sqlite_schema
		WHERE type = 'trigger' AND name NOT LIKE 'sqlite_%'
		ORDER BY name`)
	if err != nil {
		return DatabaseDocument{}, fmt.Errorf("list triggers: %w", err)
	}
	defer triggerRows.Close()
	for triggerRows.Next() {
		var trigger DatabaseTrigger
		if err := triggerRows.Scan(&trigger.Name, &trigger.Table, &trigger.SQL); err != nil {
			return DatabaseDocument{}, err
		}
		document.Triggers = append(document.Triggers, trigger)
	}
	if err := triggerRows.Err(); err != nil {
		return DatabaseDocument{}, err
	}

	document.Relationships = buildDatabaseRelationships(document.Tables)
	return document, nil
}

func verifySQLiteHeader(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	header := make([]byte, 16)
	if _, err := file.Read(header); err != nil {
		return fmt.Errorf("read SQLite header: %w", err)
	}
	if !bytes.Equal(header, []byte("SQLite format 3\x00")) {
		return fmt.Errorf("not a SQLite 3 database")
	}
	return nil
}

func objectTypeForSchema(tableListType string) string {
	if tableListType == "virtual" {
		return "table"
	}
	return tableListType
}

func sqliteSchemaSQL(database *sql.DB, objectType string, name string) (string, error) {
	var schemaSQL sql.NullString
	err := database.QueryRow(
		"SELECT sql FROM sqlite_schema WHERE type = ? AND name = ?",
		objectType, name,
	).Scan(&schemaSQL)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("read schema SQL for %s: %w", name, err)
	}
	return schemaSQL.String, nil
}

func inspectSQLiteColumns(database *sql.DB, tableName string, schemaSQL string) ([]DatabaseColumn, error) {
	rows, err := database.Query(`
		SELECT cid, name, type, "notnull", dflt_value, pk, hidden
		FROM pragma_table_xinfo(?)
		ORDER BY cid`, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns := []DatabaseColumn{}
	hasAutoIncrement := strings.Contains(strings.ToUpper(schemaSQL), "AUTOINCREMENT")
	for rows.Next() {
		var column DatabaseColumn
		var notNull, hidden int
		var defaultValue sql.NullString
		if err := rows.Scan(
			&column.Position,
			&column.Name,
			&column.Type,
			&notNull,
			&defaultValue,
			&column.PrimaryKey,
			&hidden,
		); err != nil {
			return nil, err
		}
		column.Nullable = notNull == 0 && column.PrimaryKey == 0
		if defaultValue.Valid {
			value := defaultValue.String
			column.Default = &value
		}
		column.Hidden = hidden == 1
		switch hidden {
		case 2:
			column.Generated = "virtual"
		case 3:
			column.Generated = "stored"
		}
		column.AutoIncrement = hasAutoIncrement && column.PrimaryKey > 0
		columns = append(columns, column)
	}
	return columns, rows.Err()
}

func inspectSQLiteIndexes(database *sql.DB, tableName string) ([]DatabaseIndex, error) {
	rows, err := database.Query(`
		SELECT name, "unique", origin, partial
		FROM pragma_index_list(?)
		ORDER BY seq`, tableName)
	if err != nil {
		return nil, err
	}

	indexes := []DatabaseIndex{}
	for rows.Next() {
		var index DatabaseIndex
		var unique, partial int
		if err := rows.Scan(&index.Name, &unique, &index.Origin, &partial); err != nil {
			rows.Close()
			return nil, err
		}
		index.Unique = unique != 0
		index.Partial = partial != 0
		indexes = append(indexes, index)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}

	for position := range indexes {
		index := &indexes[position]
		index.SQL, err = sqliteSchemaSQL(database, "index", index.Name)
		if err != nil {
			return nil, err
		}
		index.Columns, err = inspectSQLiteIndexColumns(database, index.Name)
		if err != nil {
			return nil, err
		}
	}
	return indexes, nil
}

func inspectSQLiteIndexColumns(database *sql.DB, indexName string) ([]DatabaseIndexColumn, error) {
	rows, err := database.Query(`
		SELECT seqno, cid, name, "desc", coll, "key"
		FROM pragma_index_xinfo(?)
		ORDER BY seqno`, indexName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns := []DatabaseIndexColumn{}
	for rows.Next() {
		var column DatabaseIndexColumn
		var name, collation sql.NullString
		var descending, key int
		if err := rows.Scan(
			&column.Position,
			&column.TableRank,
			&name,
			&descending,
			&collation,
			&key,
		); err != nil {
			return nil, err
		}
		column.Name = name.String
		column.Descending = descending != 0
		column.Collation = collation.String
		column.Key = key != 0
		columns = append(columns, column)
	}
	return columns, rows.Err()
}

func inspectSQLiteForeignKeys(database *sql.DB, tableName string) ([]DatabaseForeignKey, error) {
	rows, err := database.Query(`
		SELECT id, seq, "table", "from", "to", on_update, on_delete, match
		FROM pragma_foreign_key_list(?)
		ORDER BY id, seq`, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	foreignKeys := []DatabaseForeignKey{}
	for rows.Next() {
		var foreignKey DatabaseForeignKey
		var targetColumn, match sql.NullString
		if err := rows.Scan(
			&foreignKey.ID,
			&foreignKey.Sequence,
			&foreignKey.TargetTable,
			&foreignKey.From,
			&targetColumn,
			&foreignKey.OnUpdate,
			&foreignKey.OnDelete,
			&match,
		); err != nil {
			return nil, err
		}
		foreignKey.TargetColumn = targetColumn.String
		foreignKey.Match = match.String
		foreignKeys = append(foreignKeys, foreignKey)
	}
	return foreignKeys, rows.Err()
}

func markUniqueColumns(columns []DatabaseColumn, indexes []DatabaseIndex) {
	for _, index := range indexes {
		if !index.Unique {
			continue
		}
		var keyColumns []string
		for _, column := range index.Columns {
			if column.Key && column.Name != "" {
				keyColumns = append(keyColumns, column.Name)
			}
		}
		if len(keyColumns) != 1 {
			continue
		}
		for position := range columns {
			if columns[position].Name == keyColumns[0] {
				columns[position].Unique = true
				break
			}
		}
	}
}

func countDatabaseColumns(database DatabaseDocument) int {
	count := 0
	for _, table := range database.Tables {
		count += len(table.Columns)
	}
	for _, view := range database.Views {
		count += len(view.Columns)
	}
	return count
}

func countDatabaseRelationships(database DatabaseDocument) int {
	return len(database.Relationships)
}

func databaseDictionaryLocation(databaseSlug string, tableName string) string {
	location := "database/" + databaseSlug + "/dictionary/"
	if tableName != "" {
		location += slugify(tableName) + "/"
	}
	return location
}

func databaseDiagramLocation(databaseSlug string) string {
	return "database/" + databaseSlug + "/diagram/"
}

func buildDatabaseRelationships(tables []DatabaseTable) []DatabaseRelationship {
	relationships := []DatabaseRelationship{}
	declaredColumns := map[string]bool{}
	for _, table := range tables {
		grouped := map[int][]DatabaseForeignKey{}
		var ids []int
		for _, foreignKey := range table.ForeignKeys {
			if _, exists := grouped[foreignKey.ID]; !exists {
				ids = append(ids, foreignKey.ID)
			}
			grouped[foreignKey.ID] = append(grouped[foreignKey.ID], foreignKey)
			declaredColumns[table.Name+"\x00"+foreignKey.From] = true
		}
		sort.Ints(ids)
		for _, id := range ids {
			foreignKeys := grouped[id]
			relationship := DatabaseRelationship{
				ID:          fmt.Sprintf("%s:fk:%d", table.Name, id),
				FromTable:   table.Name,
				TargetTable: foreignKeys[0].TargetTable,
				Origin:      "declared",
				OnUpdate:    foreignKeys[0].OnUpdate,
				OnDelete:    foreignKeys[0].OnDelete,
				Evidence:    "Declared by SQLite foreign-key metadata.",
			}
			for _, foreignKey := range foreignKeys {
				relationship.FromColumns = append(relationship.FromColumns, foreignKey.From)
				relationship.TargetColumns = append(relationship.TargetColumns, foreignKey.TargetColumn)
			}
			relationships = append(relationships, relationship)
		}
	}

	for _, table := range tables {
		for _, column := range table.Columns {
			if declaredColumns[table.Name+"\x00"+column.Name] || column.PrimaryKey > 0 {
				continue
			}
			type candidate struct {
				table  DatabaseTable
				column DatabaseColumn
			}
			var candidates []candidate
			for _, targetTable := range tables {
				if targetTable.Name == table.Name {
					continue
				}
				primaryColumns := databasePrimaryColumns(targetTable)
				if len(primaryColumns) != 1 ||
					!relationshipColumnMatches(column.Name, targetTable.Name, primaryColumns[0].Name) ||
					!compatibleSQLiteTypes(column.Type, primaryColumns[0].Type) {
					continue
				}
				candidates = append(candidates, candidate{table: targetTable, column: primaryColumns[0]})
			}
			if len(candidates) != 1 {
				continue
			}
			target := candidates[0]
			relationships = append(relationships, DatabaseRelationship{
				ID:            "suggested:" + table.Name + ":" + column.Name,
				FromTable:     table.Name,
				FromColumns:   []string{column.Name},
				TargetTable:   target.table.Name,
				TargetColumns: []string{target.column.Name},
				Origin:        "suggested",
				Evidence: fmt.Sprintf(
					"Column %s uniquely matches table %s and primary key %s by name and type.",
					column.Name,
					target.table.Name,
					target.column.Name,
				),
			})
		}
	}
	sort.SliceStable(relationships, func(first int, second int) bool {
		if relationships[first].Origin != relationships[second].Origin {
			return relationships[first].Origin < relationships[second].Origin
		}
		return relationships[first].ID < relationships[second].ID
	})
	return relationships
}

func relationshipColumnMatches(sourceColumn string, targetTable string, targetPrimaryColumn string) bool {
	source := normalizeSchemaName(sourceColumn)
	target := singularSchemaName(targetTable)
	primary := normalizeSchemaName(targetPrimaryColumn)
	if source == "" || target == "" || primary == "" {
		return false
	}
	return source == primary || source == target+primary
}

func normalizeSchemaName(name string) string {
	var normalized strings.Builder
	for _, character := range strings.ToLower(name) {
		if (character >= 'a' && character <= 'z') || (character >= '0' && character <= '9') {
			normalized.WriteRune(character)
		}
	}
	return normalized.String()
}

func singularSchemaName(name string) string {
	normalized := normalizeSchemaName(name)
	if strings.HasSuffix(normalized, "ies") && len(normalized) > 3 {
		return strings.TrimSuffix(normalized, "ies") + "y"
	}
	if strings.HasSuffix(normalized, "s") && !strings.HasSuffix(normalized, "ss") && len(normalized) > 1 {
		return strings.TrimSuffix(normalized, "s")
	}
	return normalized
}

func databasePrimaryColumns(table DatabaseTable) []DatabaseColumn {
	var columns []DatabaseColumn
	for _, column := range table.Columns {
		if column.PrimaryKey > 0 {
			columns = append(columns, column)
		}
	}
	sort.Slice(columns, func(first int, second int) bool {
		return columns[first].PrimaryKey < columns[second].PrimaryKey
	})
	return columns
}

func compatibleSQLiteTypes(first string, second string) bool {
	firstAffinity := sqliteTypeAffinity(first)
	secondAffinity := sqliteTypeAffinity(second)
	return firstAffinity == "" || secondAffinity == "" || firstAffinity == secondAffinity
}

func sqliteTypeAffinity(declaredType string) string {
	upper := strings.ToUpper(strings.TrimSpace(declaredType))
	switch {
	case strings.Contains(upper, "INT"):
		return "INTEGER"
	case strings.Contains(upper, "CHAR"),
		strings.Contains(upper, "CLOB"),
		strings.Contains(upper, "TEXT"):
		return "TEXT"
	case strings.Contains(upper, "BLOB") || upper == "":
		return "BLOB"
	case strings.Contains(upper, "REAL"),
		strings.Contains(upper, "FLOA"),
		strings.Contains(upper, "DOUB"):
		return "REAL"
	default:
		return "NUMERIC"
	}
}
