# go-sqlglot Design

**Date:** 2026-03-08
**Status:** Approved

## Attribution

This project is a Go port of [sqlglot](https://github.com/tobymao/sqlglot), created by
[Toby Mao](https://github.com/tobymao) and the sqlglot contributors. This work would not
be possible without their extraordinary effort in building the original library. All credit
for the design, dialect coverage, and optimizer passes belongs to the sqlglot team.

---

## Goals

- Full feature parity with Python sqlglot: tokenizer, parser, AST, transpiler, optimizer, dialects
- Idiomatic Go API — feels natural to Go developers, not a literal Python translation
- Correctness first; performance optimization deferred until benchmarks justify it
- Initial dialect focus: **PostgreSQL, MSSQL, Microsoft Fabric, Oracle, Snowflake, Spark, DuckDB**
- Infrastructure-first approach: build the core layers, then add dialects incrementally

## Non-Goals

- Matching Python sqlglot's API surface exactly
- Supporting all ~25 dialects at launch
- High-throughput optimization in v1

---

## Architecture

### Package Layout

```
go-sqlglot/
├── tokens/
│   ├── token.go        — Token struct, TokenType enum (~150 types)
│   └── tokenizer.go    — Tokenizer, dialect-specific keyword maps
├── ast/
│   ├── node.go         — Expression interface, base node types
│   ├── expressions.go  — All concrete AST node types (Select, Join, BinaryExpr, etc.)
│   └── walk.go         — Visitor interface + Walk helper
├── parser/
│   ├── parser.go       — Base recursive-descent parser
│   └── options.go      — DialectHooks interface, parser config
├── generator/
│   ├── generator.go    — Base AST→SQL generator (strings.Builder)
│   └── options.go      — GeneratorHooks interface, generator config
├── optimizer/
│   ├── optimizer.go    — Pass runner, Schema interface
│   └── passes/
│       ├── qualify_columns.go
│       ├── pushdown_predicates.go
│       ├── eliminate_subqueries.go
│       ├── constant_fold.go
│       ├── unnest.go
│       └── eliminate_joins.go
├── dialects/
│   ├── dialect.go      — Dialect interface
│   ├── postgres/
│   ├── mssql/
│   ├── fabric/
│   ├── oracle/
│   ├── snowflake/
│   ├── spark/
│   └── duckdb/
└── sqlglot.go          — Top-level convenience API
```

### Top-Level API

```go
// Parse SQL into an AST
sqlglot.Parse(sql, sqlglot.WithDialect(postgres.Dialect))

// Transpile SQL from one dialect to another
sqlglot.Transpile(sql, from, to)

// Optimize an AST given a schema
sqlglot.Optimize(node, schema)
```

---

## Layer Design

### 1. Tokens

`TokenType` is an `int` enum (~150 values). Each dialect provides a keyword map that
overrides the base set, allowing dialect-specific keywords to tokenize differently.

```go
type Token struct {
    Type TokenType
    Text string
    Line int
    Col  int
}
```

### 2. AST

Every SQL construct implements the `Expression` interface. Concrete node types are typed
structs with explicit fields — no `interface{}` bags. This mirrors the design of Go's own
`go/ast` package, which Go developers already know.

```go
type Expression interface {
    exprNode()              // marker method
    Children() []Expression
    Copy() Expression
}
```

Example nodes:

```go
type Select struct {
    Expressions []Expression
    From        *From
    Where       *Where
    GroupBy     *Group
    Having      *Having
    OrderBy     *Order
    Limit       *Limit
    Distinct    bool
}

type BinaryExpr struct {
    Op    BinaryOp   // Eq, Neq, And, Or, Plus, ...
    Left  Expression
    Right Expression
}
```

Traversal uses a `Visitor` interface matching the pattern from `go/ast`:

```go
type Visitor interface {
    Visit(node Expression) (w Visitor, replace Expression, err error)
}

func Walk(v Visitor, node Expression) (Expression, error)
```

### 3. Parser

Recursive descent parser with one method per SQL construct. Dialects inject behavior via
a `DialectHooks` interface — the `bool` return signals whether the dialect handled the
case; if `false`, the base parser proceeds.

```go
type DialectHooks interface {
    ParseDataType(p *Parser) (ast.Expression, bool, error)
    ParseSpecialFunction(p *Parser, name string) (ast.Expression, bool, error)
    // ~20 hook points total
}
```

### 4. Generator

Walks the AST and produces a SQL string via `strings.Builder`. Dialect overrides follow
the same hooks pattern:

```go
type GeneratorHooks interface {
    GenerateDataType(g *Generator, node *ast.DataType) (string, bool, error)
    GenerateCast(g *Generator, node *ast.Cast) (string, bool, error)
    // ~20 hook points total
}
```

### 5. Dialects

Each dialect is a self-contained package implementing the `Dialect` interface. Adding a
new dialect is purely additive — no changes to core packages required.

```go
type Dialect interface {
    TokenizerConfig() tokens.Config
    ParserHooks()    parser.DialectHooks
    GeneratorHooks() generator.GeneratorHooks
}
```

### 6. Optimizer

A pipeline of composable, independently-testable rewrite passes:

```go
type Pass func(ast.Expression, *Schema) (ast.Expression, error)

type Optimizer struct {
    passes []Pass
}

func New(passes ...Pass) *Optimizer
func (o *Optimizer) Optimize(node ast.Expression, schema *Schema) (ast.Expression, error)
```

Users can run the canonical pass order via `sqlglot.Optimize()` or compose custom
pipelines from individual passes. The `Schema` interface is user-provided:

```go
type Schema interface {
    ColumnType(table, column string) (ast.DataType, bool)
    TableColumns(table string) ([]string, bool)
}
```

---

## Testing Strategy

### Fixture Tests
Ported from sqlglot's Python test suite. SQL input files and expected output files live in
`testdata/fixtures/<dialect>/`. Tests load fixtures and assert exact string output,
providing high-confidence parity with Python sqlglot behavior.

### Round-Trip Tests
Parse SQL → generate SQL → re-parse → assert AST equivalence. Catches generator bugs
that produce syntactically valid but semantically different output.

### Per-Pass Optimizer Tests
Each optimizer pass is tested in isolation with known input/output AST pairs.

### Fuzz Tests
`go test -fuzz` on the tokenizer and parser. SQL parsers are prone to panics on malformed
input; fuzzing catches these cheaply and early.

---

## Development Conventions

- **Subagent-driven development:** independent tasks are executed by parallel subagents,
  each in their own git worktree
- **PRs for all changes:** main is protected; all work lands via pull requests opened with
  the GitHub CLI
- **Reference implementation:** the Python sqlglot source is the authoritative spec for
  behavior; Go implementation may diverge in API style but must match semantics
