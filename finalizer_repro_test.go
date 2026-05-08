package googlesqlrepro

import (
	"fmt"
	"runtime"
	"testing"

	googlesql "github.com/goccy/go-googlesql"
)

func TestOwnedSimpleTableFinalizerRepro(t *testing.T) {
	if err := googlesql.Init(); err != nil {
		t.Fatalf("googlesql.Init() error = %v", err)
	}

	for i := 0; i < 1000; i++ {
		if err := buildOwnedSimpleTableCatalog(); err != nil {
			t.Fatalf("iteration %d: %v", i, err)
		}
		runtime.GC()
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
