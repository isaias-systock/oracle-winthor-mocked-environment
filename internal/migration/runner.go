package migration

import (
	"bufio"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"oracle-winthor-mocked-environment/internal/config"
	"oracle-winthor-mocked-environment/internal/db"
)

type MigrationFile struct {
	Version  string
	Name     string
	UpFile   string
	DownFile string
}

type SeedFile struct {
	Version string
	Name    string
	File    string
}

func CreateMigration(name string) error {
	if strings.TrimSpace(name) == "" {
		return errors.New("nome da migration é obrigatório")
	}

	version := time.Now().Format("20060102150405")
	base := fmt.Sprintf("%s_%s", version, sanitizeName(name))

	upPath := filepath.Join("migrations", base+".up.sql")
	downPath := filepath.Join("migrations", base+".down.sql")

	if err := os.MkdirAll("migrations", 0755); err != nil {
		return err
	}

	upContent := "-- escreva aqui o SQL de subida\n"
	downContent := "-- escreva aqui o SQL de rollback\n"

	if err := os.WriteFile(upPath, []byte(upContent), 0644); err != nil {
		return err
	}

	if err := os.WriteFile(downPath, []byte(downContent), 0644); err != nil {
		return err
	}

	fmt.Println("created:", upPath)
	fmt.Println("created:", downPath)
	return nil
}

func CreateSeed(name string) error {
	if strings.TrimSpace(name) == "" {
		return errors.New("nome da seed é obrigatório")
	}

	version := time.Now().Format("20060102150405")
	base := fmt.Sprintf("%s_%s", version, sanitizeName(name))
	path := filepath.Join("seeds", base+".sql")

	if err := os.MkdirAll("seeds", 0755); err != nil {
		return err
	}

	content := "-- escreva aqui o SQL da seed\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return err
	}

	fmt.Println("created:", path)
	return nil
}

func sanitizeName(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ReplaceAll(name, "-", "_")
	return name
}

func LoadMigrations(dir string) ([]MigrationFile, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	m := map[string]*MigrationFile{}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filename := entry.Name()

		if strings.HasSuffix(filename, ".up.sql") {
			base := strings.TrimSuffix(filename, ".up.sql")
			version, name, ok := splitMigrationBase(base)
			if !ok {
				continue
			}
			if _, exists := m[base]; !exists {
				m[base] = &MigrationFile{Version: version, Name: name}
			}
			m[base].UpFile = filepath.Join(dir, filename)
		}

		if strings.HasSuffix(filename, ".down.sql") {
			base := strings.TrimSuffix(filename, ".down.sql")
			version, name, ok := splitMigrationBase(base)
			if !ok {
				continue
			}
			if _, exists := m[base]; !exists {
				m[base] = &MigrationFile{Version: version, Name: name}
			}
			m[base].DownFile = filepath.Join(dir, filename)
		}
	}

	var result []MigrationFile
	for _, item := range m {
		if item.UpFile != "" {
			result = append(result, *item)
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Version < result[j].Version
	})

	return result, nil
}

func splitMigrationBase(base string) (version, name string, ok bool) {
	parts := strings.SplitN(base, "_", 2)
	if len(parts) != 2 {
		return "", "", false
	}
	return parts[0], parts[1], true
}

func migrationsTableName(cfg config.Config) string {
	schema := strings.TrimSpace(cfg.OracleAppUser)
	if schema == "" {
		return "MIGRATIONS"
	}
	return strings.ToUpper(schema) + ".MIGRATIONS"
}

func splitQualifiedName(name string) (schema, object string) {
	parts := strings.SplitN(strings.ToUpper(strings.TrimSpace(name)), ".", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return "", parts[0]
}

func migrationsTableExists(conn *sql.DB, tableName string) (bool, error) {
	schema, object := splitQualifiedName(tableName)

	var count int
	var err error
	if schema == "" {
		err = conn.QueryRow(
			`SELECT COUNT(1) FROM USER_TABLES WHERE TABLE_NAME = :1`,
			object,
		).Scan(&count)
	} else {
		err = conn.QueryRow(
			`SELECT COUNT(1) FROM ALL_TABLES WHERE OWNER = :1 AND TABLE_NAME = :2`,
			schema,
			object,
		).Scan(&count)
	}
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func ensureMigrationsTable(conn *sql.DB, tableName string) error {
	exists, err := migrationsTableExists(conn, tableName)
	if err != nil {
		return fmt.Errorf("ensureMigrationsTable:exists: %w", err)
	}
	if exists {
		return nil
	}

	sqlStmt := `
		BEGIN
			EXECUTE IMMEDIATE '
				CREATE TABLE ` + tableName + ` (
					VERSION VARCHAR2(32) PRIMARY KEY,
					NAME VARCHAR2(255) NOT NULL,
					TYPE VARCHAR2(20) DEFAULT ''migration'' NOT NULL,
					APPLIED_AT TIMESTAMP DEFAULT SYSTIMESTAMP NOT NULL
				)
			';
		EXCEPTION
			WHEN OTHERS THEN
				IF SQLCODE != -955 THEN
					RAISE;
				END IF;
		END;
	`
	if _, err := conn.Exec(sqlStmt); err != nil {
		return fmt.Errorf("ensureMigrationsTable:create: %w", err)
	}

	alterStmt := `
		BEGIN
			EXECUTE IMMEDIATE '
				ALTER TABLE ` + tableName + `
				ADD TYPE VARCHAR2(20) DEFAULT ''migration'' NOT NULL
			';
		EXCEPTION
			WHEN OTHERS THEN
				IF SQLCODE != -1430 THEN
					RAISE;
				END IF;
		END;
	`
	if _, err := conn.Exec(alterStmt); err != nil {
		return fmt.Errorf("ensureMigrationsTable:alter: %w", err)
	}

	_, err = conn.Exec(`UPDATE ` + tableName + ` SET TYPE = 'migration' WHERE TYPE IS NULL`)
	if err != nil {
		return fmt.Errorf("ensureMigrationsTable:update-type: %w", err)
	}
	return nil
}

func LoadSeeds(dir string) ([]SeedFile, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}

	var result []SeedFile
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filename := entry.Name()
		if !strings.HasSuffix(filename, ".sql") {
			continue
		}

		base := strings.TrimSuffix(filename, ".sql")
		version, name, ok := splitMigrationBase(base)
		if !ok {
			continue
		}

		result = append(result, SeedFile{
			Version: version,
			Name:    name,
			File:    filepath.Join(dir, filename),
		})
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Version < result[j].Version
	})

	return result, nil
}

func appliedVersions(conn *sql.DB, tableName, itemType string) (map[string]bool, error) {
	rows, err := conn.Query(`SELECT VERSION FROM `+tableName+` WHERE TYPE = :1`, itemType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make(map[string]bool)
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		out[version] = true
	}
	return out, rows.Err()
}

func markAsApplied(conn *sql.DB, tableName, version, name, itemType string) error {
	_, err := conn.Exec(
		`INSERT INTO `+tableName+` (VERSION, NAME, TYPE) VALUES (:1, :2, :3)`,
		version,
		name,
		itemType,
	)
	return err
}

func execSQLFile(conn *sql.DB, path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	statements := splitOracleStatements(string(content))
	if len(statements) == 0 {
		return nil
	}

	for i, stmt := range statements {
		if _, err := conn.Exec(stmt); err != nil {
			return fmt.Errorf("statement %d falhou: %s: %w", i+1, summarizeSQL(stmt), err)
		}
	}

	return nil
}

func splitOracleStatements(content string) []string {
	scanner := bufio.NewScanner(strings.NewReader(content))
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	var statements []string
	var current strings.Builder
	inPlSQLBlock := false

	flush := func() {
		stmt := strings.TrimSpace(current.String())
		if stmt == "" {
			current.Reset()
			return
		}
		statements = append(statements, stmt)
		current.Reset()
	}

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if trimmed == "" {
			if current.Len() > 0 {
				current.WriteString("\n")
			}
			continue
		}

		if trimmed == "/" && inPlSQLBlock {
			flush()
			inPlSQLBlock = false
			continue
		}

		if current.Len() > 0 {
			current.WriteString("\n")
		}
		current.WriteString(line)

		if !inPlSQLBlock {
			inPlSQLBlock = startsPlSQLBlock(current.String())
		}

		if inPlSQLBlock {
			continue
		}

		stmt := strings.TrimSpace(current.String())
		if strings.HasSuffix(stmt, ";") {
			stmt = strings.TrimSuffix(stmt, ";")
			current.Reset()
			current.WriteString(stmt)
			flush()
		}
	}

	flush()

	return statements
}

func startsPlSQLBlock(statement string) bool {
	normalized := normalizeSQLForDetection(statement)
	if normalized == "" {
		return false
	}

	if strings.HasPrefix(normalized, "begin ") || normalized == "begin" {
		return true
	}
	if strings.HasPrefix(normalized, "declare ") || normalized == "declare" {
		return true
	}

	if strings.HasPrefix(normalized, "create or replace ") {
		objectType := strings.TrimSpace(strings.TrimPrefix(normalized, "create or replace "))
		return strings.HasPrefix(objectType, "procedure ") ||
			strings.HasPrefix(objectType, "function ") ||
			strings.HasPrefix(objectType, "package ") ||
			strings.HasPrefix(objectType, "package body ") ||
			strings.HasPrefix(objectType, "trigger ") ||
			strings.HasPrefix(objectType, "type ") ||
			strings.HasPrefix(objectType, "type body ")
	}

	return false
}

func normalizeSQLForDetection(statement string) string {
	scanner := bufio.NewScanner(strings.NewReader(statement))
	var lines []string
	for scanner.Scan() {
		line := scanner.Text()
		if idx := strings.Index(line, "--"); idx >= 0 {
			line = line[:idx]
		}
		line = strings.TrimSpace(line)
		if line != "" {
			lines = append(lines, line)
		}
	}

	return strings.ToLower(strings.Join(lines, " "))
}

func summarizeSQL(statement string) string {
	normalized := strings.Join(strings.Fields(statement), " ")
	if len(normalized) <= 120 {
		return normalized
	}
	return normalized[:117] + "..."
}

func isOracleInsufficientPrivileges(err error) bool {
	if err == nil {
		return false
	}

	msg := strings.ToUpper(err.Error())
	return strings.Contains(msg, "ORA-01045") || strings.Contains(msg, "ORA-01031")
}

func isOracleObjectAlreadyExists(err error) bool {
	if err == nil {
		return false
	}

	msg := strings.ToUpper(err.Error())
	return strings.Contains(msg, "ORA-00955")
}

func isBootstrapMigration(m MigrationFile) bool {
	name := strings.ToLower(m.Name)
	return strings.Contains(name, "bootstrap") || strings.Contains(name, "create_winthor_schema")
}

func splitMigrationsByTarget(migrations []MigrationFile) (bootstrap []MigrationFile, app []MigrationFile) {
	for _, m := range migrations {
		if isBootstrapMigration(m) {
			bootstrap = append(bootstrap, m)
			continue
		}
		app = append(app, m)
	}
	return bootstrap, app
}

func applyMigrations(conn *sql.DB, tableName string, migrations []MigrationFile) error {
	if len(migrations) == 0 {
		return nil
	}

	if err := ensureMigrationsTable(conn, tableName); err != nil {
		return err
	}

	applied, err := appliedVersions(conn, tableName, "migration")
	if err != nil {
		return err
	}

	for _, m := range migrations {
		if applied[m.Version] {
			continue
		}

		fmt.Println("applying:", m.UpFile)

		if err := execSQLFile(conn, m.UpFile); err != nil {
			return fmt.Errorf("erro ao aplicar %s: %w", m.UpFile, err)
		}

		if err := markAsApplied(conn, tableName, m.Version, m.Name, "migration"); err != nil {
			return err
		}
	}

	return nil
}

func applySeeds(conn *sql.DB, tableName string, seeds []SeedFile) error {
	if len(seeds) == 0 {
		return nil
	}

	if err := ensureMigrationsTable(conn, tableName); err != nil {
		return err
	}

	applied, err := appliedVersions(conn, tableName, "seed")
	if err != nil {
		return err
	}

	for _, seed := range seeds {
		if applied[seed.Version] {
			continue
		}

		fmt.Println("seeding:", seed.File)

		if err := execSQLFile(conn, seed.File); err != nil {
			return fmt.Errorf("erro ao aplicar seed %s: %w", seed.File, err)
		}

		if err := markAsApplied(conn, tableName, seed.Version, seed.Name, "seed"); err != nil {
			return err
		}
	}

	return nil
}

func applyMigrationsStateless(conn *sql.DB, migrations []MigrationFile) error {
	for _, m := range migrations {
		fmt.Println("applying (stateless):", m.UpFile)

		if err := execSQLFile(conn, m.UpFile); err != nil {
			if isOracleObjectAlreadyExists(err) {
				fmt.Println("skipping existing object:", m.UpFile)
				continue
			}
			return fmt.Errorf("erro ao aplicar %s: %w", m.UpFile, err)
		}
	}

	return nil
}

func applySeedsStateless(conn *sql.DB, seeds []SeedFile) error {
	for _, seed := range seeds {
		fmt.Println("seeding (stateless):", seed.File)

		if err := execSQLFile(conn, seed.File); err != nil {
			return fmt.Errorf("erro ao aplicar seed %s: %w", seed.File, err)
		}
	}

	return nil
}

func latestAppliedMigration(conn *sql.DB, tableName, itemType string) (version string, name string, err error) {
	err = ensureMigrationsTable(conn, tableName)
	if err != nil {
		return "", "", err
	}

	err = conn.QueryRow(`
		SELECT VERSION, NAME
		FROM `+tableName+`
		WHERE TYPE = :1
		ORDER BY APPLIED_AT DESC
		FETCH FIRST 1 ROWS ONLY
	`, itemType).Scan(&version, &name)
	return version, name, err
}

func Up(cfg config.Config) error {
	migrations, err := LoadMigrations("migrations")
	if err != nil {
		return err
	}

	if len(migrations) == 0 {
		fmt.Println("nenhuma migration encontrada")
	}

	bootstrapMigrations, appMigrations := splitMigrationsByTarget(migrations)
	seeds, err := LoadSeeds("seeds")
	if err != nil {
		return err
	}

	if len(bootstrapMigrations) == 0 && len(appMigrations) == 0 && len(seeds) == 0 {
		fmt.Println("migrations aplicadas com sucesso!")
		return nil
	}

	adminConn, err := db.OpenAdmin(cfg)
	if err != nil {
		return fmt.Errorf("falha ao abrir conexão admin para migrations: %w", err)
	}
	defer adminConn.Close()

	if len(bootstrapMigrations) > 0 {
		if err := applyMigrationsStateless(adminConn, bootstrapMigrations); err != nil {
			return err
		}
	}

	if err := applyMigrationsStateless(adminConn, appMigrations); err != nil {
		return err
	}

	if err := applySeedsStateless(adminConn, seeds); err != nil {
		return err
	}

	fmt.Println("migrations aplicadas com sucesso!")
	return nil
}

func Down(cfg config.Config) error {
	migrations, err := LoadMigrations("migrations")
	if err != nil {
		return err
	}

	tableName := migrationsTableName(cfg)
	adminConn, err := db.OpenAdmin(cfg)
	if err != nil {
		return fmt.Errorf("falha ao abrir conexão admin para rollback: %w", err)
	}
	defer adminConn.Close()

	version, _, err := latestAppliedMigration(adminConn, tableName, "migration")
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("nenhuma migration aplicada para rollback")
		}
		return err
	}

	for _, m := range migrations {
		if m.Version == version {
			if m.DownFile == "" {
				return fmt.Errorf("migration %s não possui arquivo down", m.Name)
			}

			fmt.Println("rolling back bootstrap:", m.DownFile)

			if err := execSQLFile(adminConn, m.DownFile); err != nil {
				return err
			}

			_, err := adminConn.Exec(`DELETE FROM `+tableName+` WHERE VERSION = :1 AND TYPE = 'migration'`, version)
			return err
		}
	}

	return fmt.Errorf("migration aplicada não encontrada em disco: %s", version)
}
