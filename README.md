# cssx

cssx is a tiny, standalone Go library that parses extended CSS expressions and compiles them into a linear postfix pipeline IR.

It does not execute anything and does not depend on a browser, CDP, or DOM. It only provides:

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
	pipeline, err := cssx.Compile(`:attr("href", :first(a.cta))`)
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

- selector steps (`Select`)
- call steps (`Call`)
- literal call args embedded in each `Call.Args`

Example:

Input:

```
:attr("href", :first(a.cta))
```

Pipeline:

```
[
  Select("a.cta"),
  Call("first", Arity:1, Args:[]),
  Call("attr", Arity:1, Args:[String("href")]),
]
```

## Syntax

### 1) Plain CSS selector mode

If input does not start with `:`, it is treated as a plain selector.

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

- `>>` splits pipeline only at top level (not inside quotes/brackets/parens)
- First segment: plain selector (must not start with `:`)
- Each following segment: must be a call `:name(...)`
- Zero-arg calls must still include parentheses: `:text()`

Compiles to:

```
[
  Select(".section .item"),
  Call("nth", Arity:0, Args:[Number(2)]),
  Call("text", Arity:0, Args:[]),
]
```

## IR Contract

### `Arity`

`Op.Arity` is the number of non-literal expression args consumed from the stack.

### `Args`

`Op.Args` contains only literal arguments (`string` or `number`) in source order.

### `literals-first` rule

Literal args must come before expression args inside a call.

Valid:

- `:foo("x", 2, :bar(a))`

Invalid:

- `:foo(:bar(a), "x")`
- `:foo(1, :bar(a), 2)`

Invalid calls fail at compile time with:

- `literal args must precede expression args`

## Supported Examples

- `.product`
  - `[Select(".product")]`
- `:count(.product)`
  - `[Select(".product"), Call("count", Arity:1, Args:[])]`
- `:text(:first(h1))`
  - `[Select("h1"), Call("first", Arity:1, Args:[]), Call("text", Arity:1, Args:[])]`
- `:attr("href", :first(a.cta))`
  - `[Select("a.cta"), Call("first", Arity:1, Args:[]), Call("attr", Arity:1, Args:[String("href")])]`
- `:first(a[href*="x,y"])`
  - `[Select("a[href*=\"x,y\"]"), Call("first", Arity:1, Args:[])]`
- `:text(:nth(2, .section .item))`
  - `[Select(".section .item"), Call("nth", Arity:1, Args:[Number(2)]), Call("text", Arity:1, Args:[])]`
- `.section .item >> :nth(2) >> :text()`
  - `[Select(".section .item"), Call("nth", Arity:0, Args:[Number(2)]), Call("text", Arity:0, Args:[])]`

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
	Args     []CallArg
}

type CallArg struct {
	Kind CallArgKind
	Str  string
	Num  float64
}

type ParseError struct {
	Message string
	Pos     int // byte offset
}
```

## Error Handling

All syntax errors and compile-time IR validation errors return `*ParseError` with a byte offset position.

Common error cases:

- `:` (missing identifier)
- `:text(` (unterminated call)
- `:text(:first(h1)` (missing `)`) 
- `:attr("href" :first(a))` (missing comma)
- empty input (whitespace only)
- pipeline stage without `:name(...)`
- mixed call arg order violating literals-first

## Non-Goals

- No DOM, JS, or runtime execution
- No CSS selector validation
- No pseudo validation (names and arity are not checked)
- No external dependencies

## License

See `LICENSE`.
