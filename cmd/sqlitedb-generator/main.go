package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/tshiiba/sqlitedb-generator/internal/generator"

	_ "modernc.org/sqlite"
)

func main() {
	var (
		inDir     = flag.String("in", "./tsv", "input directory containing .tsv files")
		outPath   = flag.String("out", "./out.db", "output SQLite database file path")
		overwrite = flag.Bool("overwrite", false, "overwrite output db file if it already exists")
		drop      = flag.Bool("drop", false, "drop and recreate tables before importing")
		verbose   = flag.Bool("v", false, "verbose output")
	)
	flag.Parse()

	absIn, err := filepath.Abs(*inDir)
	if err != nil {
		fatal(err)
	}
	absOut, err := filepath.Abs(*outPath)
	if err != nil {
		fatal(err)
	}

	if !*overwrite {
		if _, err := os.Stat(absOut); err == nil {
			fatal(fmt.Errorf("output db already exists (use -overwrite): %s", absOut))
		}
	}
	if *overwrite {
		_ = os.Remove(absOut)
	}

	db, err := sql.Open("sqlite", absOut)
	if err != nil {
		fatal(err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "warning: close db error: %v\n", err)
		}
	}()

	// Make sure we can talk to SQLite.
	if err := db.Ping(); err != nil {
		fatal(err)
	}

	ctx := context.Background()
	if _, err := db.ExecContext(ctx, "PRAGMA journal_mode=WAL"); err != nil {
		// Not fatal; continue.
		if *verbose {
			fmt.Fprintf(os.Stderr, "warning: failed to set WAL: %v\n", err)
		}
	}

	err = generator.Run(ctx, db, generator.Options{
		InputDir:  absIn,
		OutputDB:  absOut,
		Overwrite: *overwrite,
		DropTable: *drop,
		Verbose:   *verbose,
	})
	if err != nil {
		fatal(err)
	}

	if *verbose {
		fmt.Printf("done: %s\n", absOut)
	}
}

func fatal(err error) {
	if err == nil {
		return
	}
	if errors.Is(err, flag.ErrHelp) {
		os.Exit(2)
	}
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
