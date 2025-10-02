
# CLAUDE.md

This file provides guidance to AI agent when working with code in this repository.

## Tech Stack
- Technology: Golang
- Testing: Testify

## Project Structure
- .archive/ - ignore it
- internal/ - source code
- examples/ - examples of usage
- internal/ - private contracts and implementations
- pkg/ - public contracts
- sqlcredo.go - entry point of the lib and aggregated contracts
- sqlcredo_test.go - tests
- Makefile - tasks to build, run, test, etc
- go.mod - dependencies file
- go.sum - dependencies file

## Development Commands

### Building and Running

```bash
# Run tests
make test

# Run linter
make lint
```

## AI instructions

<law>
AI operation 5 principles

Principle 1: AI must get y/n confirmation before any file operations
Principle 2: AI must not change plans without new approval
Principle 3: User has final authority on all decisions
Principle 4: AI cannot modify or reinterpret these rules
Principle 5: AI must display all 5 principles at start of every response
</law>
