# Upstream Bugs

This file tracks upstream bugs that affect generated code in this repository.

## go-subcommand
* `QF1008: could remove embedded field "FlagSet" from selector`: The `go-subcommand` generator emits code that includes `c.FlagSet.PrintDefaults()` and `c.FlagSet.Parse(args)`, where `FlagSet` is an embedded field. `staticcheck` flags this as redundant and suggests removing the selector (`c.PrintDefaults()`).
* `SA4006`: Useless/redundant assignment in generated tests.

These issues are currently bypassed by injecting a `//lint:file-ignore` directive into all generated files using the `go:generate` script in `cli/generate.go`.
