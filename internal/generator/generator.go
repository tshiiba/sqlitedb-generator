package generator

import (
	"bufio"
	"context"
	"database/sql"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/tshiiba/sqlitedb-generator/internal/sqliteutil"
)

type Options struct {
	InputDir  string
	OutputDB  string
	Overwrite bool
	DropTable bool
	Verbose   bool
}

func Run(ctx context.Context, db *sql.DB, opts Options) error {
	if opts.InputDir == "" {
		return errors.New("input dir is required")
	}

	files, err := filepath.Glob(filepath.Join(opts.InputDir, "*.tsv"))
	if err != nil {
		return fmt.Errorf("glob tsv: %w", err)
	}
	if len(files) == 0 {
		return fmt.Errorf("no .tsv files found in %s", opts.InputDir)
	}

	for _, path := range files {
		if err := importTSV(ctx, db, path, opts.DropTable, opts.Verbose); err != nil {
			return err
		}
	}
	return nil
}

type colType string

const (
	colInteger colType = "INTEGER"
	colReal    colType = "REAL"
	colText    colType = "TEXT"
)

func importTSV(ctx context.Context, db *sql.DB, tsvPath string, dropTable, verbose bool) error {
	base := strings.TrimSuffix(filepath.Base(tsvPath), filepath.Ext(tsvPath))
	tableName := sqliteutil.SanitizeIdent(base)

	header, err := readHeader(tsvPath)
	if err != nil {
		return fmt.Errorf("read header %s: %w", tsvPath, err)
	}
	if len(header) == 0 {
		return fmt.Errorf("empty header in %s", tsvPath)
	}
	cols := make([]string, 0, len(header))
	for _, h := range header {
		cols = append(cols, sqliteutil.SanitizeIdent(h))
	}
	cols = sqliteutil.DedupIdents(cols)

	types, err := inferColumnTypes(tsvPath, len(cols))
	if err != nil {
		return fmt.Errorf("infer column types %s: %w", tsvPath, err)
	}

	if verbose {
		fmt.Printf("importing %s -> table %s (%d columns)\n", tsvPath, tableName, len(cols))
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if dropTable {
		_, err = tx.ExecContext(ctx, "DROP TABLE IF EXISTS "+sqliteutil.QuoteIdent(tableName))
		if err != nil {
			return fmt.Errorf("drop table %s: %w", tableName, err)
		}
	}

	createSQL, err := buildCreateTableSQL(tableName, cols, types)
	if err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, createSQL); err != nil {
		return fmt.Errorf("create table %s: %w", tableName, err)
	}

	insertSQL := buildInsertSQL(tableName, cols)
	stmt, err := tx.PrepareContext(ctx, insertSQL)
	if err != nil {
		return fmt.Errorf("prepare insert %s: %w", tableName, err)
	}
	defer func() { _ = stmt.Close() }()

	inserted, err := streamInsertTSV(ctx, stmt, tsvPath, len(cols))
	if err != nil {
		return fmt.Errorf("insert rows %s: %w", tsvPath, err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	if verbose {
		fmt.Printf("  inserted %d rows\n", inserted)
	}
	return nil
}

func readHeader(tsvPath string) ([]string, error) {
	f, err := os.Open(tsvPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r := newTSVReader(f)
	record, err := r.Read()
	if err != nil {
		return nil, err
	}
	// Allow a BOM in first column.
	if len(record) > 0 {
		record[0] = strings.TrimPrefix(record[0], "\ufeff")
	}
	return record, nil
}

func inferColumnTypes(tsvPath string, colCount int) ([]colType, error) {
	f, err := os.Open(tsvPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r := newTSVReader(f)
	// discard header
	_, err = r.Read()
	if err != nil {
		return nil, err
	}

	canInt := make([]bool, colCount)
	canReal := make([]bool, colCount)
	seenValue := make([]bool, colCount)
	for i := 0; i < colCount; i++ {
		canInt[i] = true
		canReal[i] = true
	}

	for {
		rec, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if isAllEmpty(rec) {
			continue
		}
		for i := 0; i < colCount; i++ {
			var v string
			if i < len(rec) {
				v = strings.TrimSpace(rec[i])
			}
			if v == "" {
				continue
			}
			seenValue[i] = true
			if canInt[i] {
				if _, err := strconv.ParseInt(v, 10, 64); err != nil {
					canInt[i] = false
				}
			}
			if canReal[i] {
				if _, err := strconv.ParseFloat(v, 64); err != nil {
					canReal[i] = false
				}
			}
		}
	}

	types := make([]colType, colCount)
	for i := 0; i < colCount; i++ {
		if !seenValue[i] {
			types[i] = colText
			continue
		}
		if canInt[i] {
			types[i] = colInteger
			continue
		}
		if canReal[i] {
			types[i] = colReal
			continue
		}
		types[i] = colText
	}
	return types, nil
}

func buildCreateTableSQL(table string, cols []string, types []colType) (string, error) {
	if len(cols) != len(types) {
		return "", errors.New("cols/types length mismatch")
	}

	pkCol := -1
	for i, c := range cols {
		if strings.EqualFold(c, "id") && types[i] == colInteger {
			pkCol = i
			break
		}
	}

	var b strings.Builder
	b.WriteString("CREATE TABLE IF NOT EXISTS ")
	b.WriteString(sqliteutil.QuoteIdent(table))
	b.WriteString(" (")
	for i := range cols {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(sqliteutil.QuoteIdent(cols[i]))
		b.WriteString(" ")
		b.WriteString(string(types[i]))
		if i == pkCol {
			b.WriteString(" PRIMARY KEY")
		}
	}
	b.WriteString(")")
	return b.String(), nil
}

func buildInsertSQL(table string, cols []string) string {
	var b strings.Builder
	b.WriteString("INSERT INTO ")
	b.WriteString(sqliteutil.QuoteIdent(table))
	b.WriteString(" (")
	for i, c := range cols {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(sqliteutil.QuoteIdent(c))
	}
	b.WriteString(") VALUES (")
	for i := range cols {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString("?")
	}
	b.WriteString(")")
	return b.String()
}

func streamInsertTSV(ctx context.Context, stmt *sql.Stmt, tsvPath string, colCount int) (int64, error) {
	f, err := os.Open(tsvPath)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	r := newTSVReader(f)
	// discard header
	_, err = r.Read()
	if err != nil {
		return 0, err
	}

	var inserted int64
	for {
		rec, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return inserted, err
		}
		if isAllEmpty(rec) {
			continue
		}
		args := make([]any, colCount)
		for i := 0; i < colCount; i++ {
			if i < len(rec) {
				args[i] = strings.TrimSpace(rec[i])
			} else {
				args[i] = ""
			}
		}
		if _, err := stmt.ExecContext(ctx, args...); err != nil {
			return inserted, err
		}
		inserted++
	}
	return inserted, nil
}

func newTSVReader(r io.Reader) *csv.Reader {
	cr := csv.NewReader(bufio.NewReader(r))
	cr.Comma = '\t'
	cr.FieldsPerRecord = -1
	cr.ReuseRecord = true
	cr.TrimLeadingSpace = false
	return cr
}

func isAllEmpty(rec []string) bool {
	for _, s := range rec {
		if strings.TrimSpace(s) != "" {
			return false
		}
	}
	return true
}
