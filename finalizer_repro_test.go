package googlesqlrepro

import (
	"fmt"
	"os"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"testing"

	googlesql "github.com/goccy/go-googlesql"
)

const (
	disableGCEnv  = "GO_GOOGLESQL_REPRO_DISABLE_GC"
	iterationsEnv = "GO_GOOGLESQL_REPRO_ITERATIONS"
)

var disableGC bool

func TestMain(m *testing.M) {
	disableGC = disableGC || truthy(os.Getenv(disableGCEnv))
	if !disableGC {
		os.Exit(m.Run())
	}

	oldGCPercent := debug.SetGCPercent(-1)
	code := m.Run()
	debug.SetGCPercent(oldGCPercent)
	os.Exit(code)
}

func TestOwnedSimpleTableFinalizerRepro(t *testing.T) {
	if err := googlesql.Init(); err != nil {
		t.Fatalf("googlesql.Init() error = %v", err)
	}

	for i := 0; i < reproIterations(t); i++ {
		if err := buildOwnedSimpleTableCatalog(); err != nil {
			t.Fatalf("iteration %d: %v", i, err)
		}
		maybeGC()
	}
}

func reproIterations(t *testing.T) int {
	t.Helper()
	if raw := strings.TrimSpace(os.Getenv(iterationsEnv)); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n < 1 {
			t.Fatalf("%s must be a positive integer, got %q", iterationsEnv, raw)
		}
		return n
	}
	return 1000
}

func maybeGC() {
	if disableGC {
		return
	}
	runtime.GC()
}

func truthy(raw string) bool {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "1", "t", "true", "y", "yes", "on":
		return true
	default:
		return false
	}
}

func buildOwnedSimpleTableCatalog() error {
	tf, err := googlesql.NewTypeFactory()
	if err != nil {
		return fmt.Errorf("NewTypeFactory: %w", err)
	}

	catalog, err := googlesql.NewSimpleCatalog("spanner", tf)
	if err != nil {
		return fmt.Errorf("NewSimpleCatalog: %w", err)
	}

	lang, err := googlesql.NewLanguageOptions()
	if err != nil {
		return fmt.Errorf("NewLanguageOptions: %w", err)
	}
	if err := lang.EnableMaximumLanguageFeaturesForDevelopment(); err != nil {
		return fmt.Errorf("EnableMaximumLanguageFeaturesForDevelopment: %w", err)
	}
	if err := lang.SetSupportsAllStatementKinds(); err != nil {
		return fmt.Errorf("SetSupportsAllStatementKinds: %w", err)
	}
	if err := lang.SetProductMode(googlesql.ProductModeProductExternal); err != nil {
		return fmt.Errorf("SetProductMode: %w", err)
	}
	if err := catalog.AddBuiltinFunctionsAndTypes(&googlesql.BuiltinFunctionOptions{LanguageOptions: lang}); err != nil {
		return fmt.Errorf("AddBuiltinFunctionsAndTypes: %w", err)
	}

	opts, err := googlesql.NewAnalyzerOptions2()
	if err != nil {
		return fmt.Errorf("NewAnalyzerOptions2: %w", err)
	}
	if err := opts.SetLanguage(lang); err != nil {
		return fmt.Errorf("SetLanguage: %w", err)
	}

	table, err := googlesql.NewSimpleTable("Singers", -1)
	if err != nil {
		return fmt.Errorf("NewSimpleTable: %w", err)
	}
	if err := table.SetAllowDuplicateColumnNames(true); err != nil {
		return fmt.Errorf("SetAllowDuplicateColumnNames: %w", err)
	}
	if err := table.SetAllowAnonymousColumnName(true); err != nil {
		return fmt.Errorf("SetAllowAnonymousColumnName: %w", err)
	}

	int64Type, err := tf.GetInt64()
	if err != nil {
		return fmt.Errorf("GetInt64: %w", err)
	}
	column, err := googlesql.NewSimpleColumn("Singers", "SingerId", int64Type, false, false)
	if err != nil {
		return fmt.Errorf("NewSimpleColumn: %w", err)
	}
	if err := table.AddColumn2(column, true); err != nil {
		return fmt.Errorf("AddColumn2: %w", err)
	}
	if err := table.SetPrimaryKey([]int32{0}); err != nil {
		return fmt.Errorf("SetPrimaryKey: %w", err)
	}
	if err := catalog.AddOwnedTable(table); err != nil {
		return fmt.Errorf("AddOwnedTable: %w", err)
	}

	return nil
}
