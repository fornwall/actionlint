Go API
======
[![API Document][api-badge]][apidoc]

This document describes how to use [actionlint](..) as Go library.

actionlint can be used from Go programs. See [the documentation][apidoc] to know the list of all APIs. It contains
a workflow file parser built on top of `go-yaml/yaml`, expression `${{ }}` lexer/parser/checker, etc.

Followings are unexhaustive list of interesting APIs.

- `Command` struct represents entire `actionlint` command. `Command.Main` takes command line arguments and runs command
  until the end and returns exit status.
- `Linter` manages linter lifecycle and applies checks to given files. If you want to run actionlint checks in your
  program, please use this struct.
- `Project` and `Projects` detect a project (Git repository) in a given directory path and find configuration in it.
- `Config` represents structure of `actionlint.yaml` config file. It can be decoded by [go-yaml/yaml][go-yaml] library.
- `Workflow`, `Job`, `Step`, ... are nodes of workflow syntax tree. `Workflow` is a root node.
- `Parse()` parses given contents into a workflow syntax tree. It tries to find syntax errors as much as possible and
  returns found errors as slice.
- `Pass` is a visitor to traverse a workflow syntax tree. Multiple passes can be applied at single pass using `Visitor`.
- `Rule` is an interface for rule checkers and `RuneBase` is a base struct to implement a rule checker.
  - `RuleExpression` is a rule checker to check expression syntax in `${{ }}`.
  - `RuleShellcheck` is a rule checker to apply `shellcheck` command to `run:` sections and collect errors from it.
  - `RuleJobNeeds` is a rule checker to check dependencies in `needs:` section. It can detect cyclic dependencies.
  - ...
- `ExprLexer` lexes expression syntax in `${{ }}` and returns slice of `Token`.
- `ExprParser` parses given slice of `Token` and returns syntax tree for expression in `${{ }}`. `ExprNode` is an
  interface for nodes in the expression syntax tree.
- `ExprType` is an interface of types in expression syntax `${{ }}`. `ObjectType`, `ArrayType`, `StringType`,
  `NumberType`, ... are structs to represent actual types of expression.
- `ExprSemanticsChecker` checks semantics of expression syntax `${{ }}`. It traverses given expression syntax tree and
  deduces its type, checking types and resolving variables (contexts).
- `ValidateRefGlob()` and `ValidatePathGlob()` validate [glob filter pattern][filter-pattern-doc] and returns all errors
  found by the validator.
- `ActionMetadata` is a struct for action metadata file (`action.yml`). It is used to check inputs specified at `with:`
  and typing `steps.{id}.outputs` object strictly.
- `PopularActions` global variable is data set of popular actions' metadata collected by [the script](../scripts/generate-popular-actions).

Note that the version of this repository is for command line tool `actionlint`. So it does not represent version of the
library. It means that patch version bump may introduce some breaking changes.

---

[Checks](checks.md) | [Installation](install.md) | [Usage](usage.md) | [Configuration](config.md) | [References](reference.md)

[api-badge]: https://pkg.go.dev/badge/github.com/rhysd/actionlint.svg
[apidoc]: https://pkg.go.dev/github.com/rhysd/actionlint
[go-yaml]: https://github.com/go-yaml/yaml
[filter-pattern-doc]: https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#filter-pattern-cheat-sheet
