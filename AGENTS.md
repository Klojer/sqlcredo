# AGENTS.md - Development Guidelines for sqlcredo

## Build/Lint/Test Commands
- **Run all tests**: `make test` or `go test -v -race -shuffle=on ./...`
- **Run single test**: `go test -v -run TestName ./path/to/package`
- **Test coverage**: `make test-cov`
- **Lint**: `make lint` or `golangci-lint run ./...`
- **Dependencies**: `make dependencies` or `go mod tidy`
- **Install git hooks**: `make install-git-hooks`

## Code Style Guidelines

### Imports
- Standard library imports first
- Third-party packages second
- Internal packages last
- Group with blank lines between sections

### Naming Conventions
- Exported: PascalCase (functions, types, constants)
- Unexported: camelCase (functions, variables)
- Acronyms: Consistent casing (SQLCredo, not SqlCredo)

### Types and Generics
- Use generics extensively: `[T any, I comparable]`
- Define interfaces for contracts
- Use struct tags: `db:"field_name"`

### Error Handling
- Wrap errors: `fmt.Errorf("context: %w", err)`
- Return early on errors
- Use domain-specific error contexts

### Comments
- Godoc comments for all exported functions/types
- Include parameter/return descriptions
- Explain purpose and usage

### Testing
- Use testify (assert, require)
- Table-driven tests for multiple cases
- Test helpers with t.Cleanup()
- Mock external dependencies when needed

### Constants
- Group related constants in const blocks
- Use meaningful names with context

### Code Organization
- internal/ for private implementations
- examples/ for usage demonstrations
- Clear separation of concerns</content>
<parameter name="filePath">/home/cora/soul/src/sqlcredo/AGENTS.md