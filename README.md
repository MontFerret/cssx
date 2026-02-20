# cssx

cssx is a tiny, standalone Go library that parses **extended CSS expressions** and compiles them into a linear **postfix pipeline IR**.

It **does not execute** anything and **does not depend** on a browser, CDP, or DOM. It only provides:

- Parser: `input string -> AST`
- Builder: `AST -> Pipeline IR`

A separate package can then translate this IR into JS or any other runtime.

## Installation

```bash
go get github.com/MontFerret/cssx
```

## Quick Start

```go
package main

import (
    "fmt"

    "github.com/MontFerret/cssx"
)

func main() {
    pipeline, err := cssx.Compile(":text(:first(h1))")
    if err != nil {
        panic(err)
    }

    for _, op := range pipeline.Ops {
        fmt.Printf("%s: %+v\n", op.Kind, op)
    }
}
```

## Core Idea

Humans write nested pseudo calls like:

- `:count(.product)`
- `:text(:first(h1))`
- `:attr("href", :first(a.cta))`

cssx compiles them into a linear postfix pipeline:

- native CSS selector steps (`Select`)
- pseudo/function steps (`Call`)
- constants (`Str`, `Num`)

Example:

Input:

```
:text(:first(h1))
```

Pipeline:

```
[Select("h1"), Call("first",1), Call("text",1)]
```

## Syntax

### 1) Plain CSS selector mode

If input does **not** start with `:`, it is treated as a plain selector.

```
.section .item
```

Compiles to:

```
[Select(".section .item")]
```

### 2) Expression mode (starts with `:`)

```
:text(:nth(2, .section .item))
```

Rules:

- `Expr := SelectorExpr | CallExpr | StringLit | NumberLit`
- `CallExpr := ':' Ident '(' ArgList? ')'`
- `ArgList := Arg (',' Arg)*`
- `SelectorExpr` is raw selector text until `,` or `)` at the current nesting depth
- Strings support `"..."` and `'...'` with escapes
- Numbers are parsed with `strconv.ParseFloat`

### 3) Pipeline sugar (explicit delimiter)

```
.section .item >> :nth(2) >> :text()
```

Rules:

- `>>` splits pipeline **only at top level** (not inside quotes/brackets/parens)
- First segment: **plain selector** (must not start with `:`)
- Each following segment: **must be a call** `:name(...)`
- Zero-arg calls must still include parentheses: `:text()`

Compiles to:

```
[Select(".section .item"), Num(2), Call("nth",1), Call("text",0)]
```

## Supported Examples

Plain selectors:

- `.product` -> `[Select(".product")]`
- `.section .item` -> `[Select(".section .item")]`

Single pseudo:

- `:count(.product)` -> `[Select(".product"), Call("count",1)]`
- `:first(section)` -> `[Select("section"), Call("first",1)]`

Nested pseudos:

- `:text(:first(h1))` -> `[Select("h1"), Call("first",1), Call("text",1)]`
- `:attr("href", :first(a.cta))` -> `[Str("href"), Select("a.cta"), Call("first",1), Call("attr",2)]`

Selectors with commas inside:

- `:first(a[href*="x,y"])` -> `[Select("a[href*="x,y"]"), Call("first",1)]`

Mixed readability:

- `:text(:nth(2, .section .item))` -> `[Num(2), Select(".section .item"), Call("nth",2), Call("text",1)]`

Pipeline sugar:

- `.section .item >> :nth(2) >> :text()` -> `[Select(".section .item"), Num(2), Call("nth",1), Call("text",0)]`

## Public API

```go
package cssx

// Parse parses input into an AST.
func Parse(input string) (AST, error)

// ParseToAST is an alias for Parse.
func ParseToAST(input string) (AST, error)

// BuildPipeline compiles an AST into a postfix pipeline IR.
func BuildPipeline(ast AST) (Pipeline, error)

// Compile parses input and builds a pipeline.
func Compile(input string) (Pipeline, error)
```

Key types:

```go
type Pipeline struct {
    Ops []Op
}

type Op struct {
    Kind     OpKind
    Selector string
    Name     string
    Arity    int
    Str      string
    Num      float64
}

type ParseError struct {
    Msg string
    Pos int // byte offset
}
```

## Error Handling

All syntax errors return `*ParseError` with a byte offset position.

Common error cases:

- `:` (missing identifier)
- `:text(` (unterminated call)
- `:text(:first(h1)` (missing `)`)
- `:attr("href" :first(a))` (missing comma)
- empty input (whitespace only)
- pipeline stage without `:name(...)`

## Non-Goals

- No DOM, JS, or runtime execution
- No CSS selector validation
- No pseudo validation (names and arity are not checked)
- No external dependencies

## License

See `LICENSE`.
