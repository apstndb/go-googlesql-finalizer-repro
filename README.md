# go-googlesql finalizer repro

This repository is a minimal reproduction for a `github.com/goccy/go-googlesql`
v0.2.0 WASM memory error that appears when short-lived owned catalog objects are
created and then Go finalizers are allowed to run.

The test repeatedly builds a `SimpleCatalog`, adds built-in functions and types,
creates a `SimpleTable` with an owned `SimpleColumn`, adds it with
`AddOwnedTable`, and then calls `runtime.GC()`.

## Reproduce

```sh
go test -count=1 -v ./...
```

The failure is timing-sensitive, so the test loops enough times to make the
problem likely to reproduce. A run may take a couple of minutes.

## Observed Failure

Observed with `github.com/goccy/go-googlesql v0.2.0`:

```text
=== RUN   TestOwnedSimpleTableFinalizerRepro
    finalizer_repro_test.go:18: iteration 316: AddBuiltinFunctionsAndTypes: wasm export call: wasm error: out of bounds memory access
        wasm stack trace:
        	.$39229(i32) i32
        	.$39128(i32) i32
        	.$18125(i32,i32,i32,i64) i32
        	.$18079(i32,i32,i32,i32,i32,i32)
        	.$35108(i32,i32,i32,i32)
        	.$35110(i32,i32,i32)
        	.$14258(i32,i32) i64
--- FAIL: TestOwnedSimpleTableFinalizerRepro (112.95s)
FAIL
FAIL	github.com/apstndb/go-googlesql-finalizer-repro	114.080s
FAIL
```

In a larger analyzer test suite, the same pattern also shows up later in
unrelated tests as `wasm error: out of bounds memory access`, `wasm error:
unreachable`, or `wasm error: invalid table access`.
