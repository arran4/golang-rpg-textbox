# Upstream Bugs

* [Fix upstream bugs QF1008 and SA4006 in generated code](https://github.com/arran4/go-subcommand/pull/396)
* Generating templates using v0.0.20 introduces a right trim marker missing a space (`*/-}}`) which causes a Go template parsing panic: `comment ends before closing delimiter`.
