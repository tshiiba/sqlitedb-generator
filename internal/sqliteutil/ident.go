package sqliteutil

import (
	"fmt"
	"strings"
	"unicode"
)

// QuoteIdent returns an SQLite identifier quoted with double quotes.
// It also escapes any embedded double quotes.
func QuoteIdent(ident string) string {
	return "\"" + strings.ReplaceAll(ident, "\"", "\"\"") + "\""
}

// SanitizeIdent makes a best-effort SQLite identifier from an arbitrary string.
// It replaces non [A-Za-z0-9_] with '_' and ensures the result is non-empty.
func SanitizeIdent(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "col"
	}

	var b strings.Builder
	b.Grow(len(raw))
	for _, r := range raw {
		if r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
			continue
		}
		b.WriteByte('_')
	}
	out := b.String()
	out = strings.Trim(out, "_")
	if out == "" {
		return "col"
	}
	// SQLite identifiers can start with a digit if quoted, but we still normalize
	// to make generated SQL easier to read.
	if out[0] >= '0' && out[0] <= '9' {
		out = "c_" + out
	}
	return out
}

func DedupIdents(idents []string) []string {
	seen := make(map[string]int, len(idents))
	out := make([]string, 0, len(idents))
	for _, ident := range idents {
		base := ident
		n := seen[base]
		if n == 0 {
			seen[base] = 1
			out = append(out, base)
			continue
		}
		for {
			n++
			cand := fmt.Sprintf("%s_%d", base, n)
			if _, ok := seen[cand]; ok {
				continue
			}
			seen[base] = n
			seen[cand] = 1
			out = append(out, cand)
			break
		}
	}
	return out
}
