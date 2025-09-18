## Contributing to go-generics

Thank you for your interest in contributing to go-generics! This document provides guidelines for contributing to this project.

## Table of Contents
- [Development Setup](#development-setup)
- [Making Changes](#making-changes)
- [Testing](#testing)
- [Code Style](#code-style)
- [Pull Request Process](#pull-request-process)

## Development Setup

### Prerequisites
- Go 1.24.0 or later
- Docker or Podman (for Redis integration tests)
- Make

## Making Changes

### Before You Start
- Check existing issues and PRs to avoid duplicating work
- Create an issue for major changes to discuss the approach
- Follow existing code patterns and conventions

### Code Guidelines
- Write generic, reusable code that follows Go best practices
- Include comprehensive documentation with examples
- Ensure your code has no instances of exponential time complexity
- Avoid n+1 network request patterns
- Use meaningful variable and function names

### Package Organization
The project is organized into focused packages:
- `functions/` - Generic utility functions with concurrency support
- `datastructures/` - Generic data structures (Set, Stack, etc.)
- `cache/` - Generic caching implementations
- `ratelimit/` - Rate limiting functionality
- `concurrency/` - Concurrency utilities and distributed locking
- `encryption/` - Generic encryption/decryption utilities
- `serialize/` - Serialization helpers

## Testing

### Running Tests
```bash
# Run all unit tests
make test

# Run tests with coverage
make test-coverage

# Run fuzz tests (30 seconds)
make test-fuzz

# Start Redis and run Redis integration tests
make test-redis

```

### Test Requirements
- All new code must include comprehensive tests
- Tests should use the `//go:build test` build tag
- Use the `internal/test` package helpers for assertions
- Include both unit tests and integration tests where applicable
- Test edge cases and error conditions
- Aim for high test coverage

### Test Structure

```go
//go:build test
// +build test

package yourpackage

import (
    "testing"
    "github.com/zendesk/go-generics/internal/test"
)

func TestYourFunction(t *testing.T) {
    // Use test.CheckOk, test.CheckEqual, etc.
}
```

## Code Style

### Formatting
- Use `gofmt` for code formatting: `make fmt`
- Follow standard Go naming conventions
- Use meaningful comments for exported functions and types

### Documentation
- Update README.md for new features

## Pull Request Process

### Before Submitting
1. Ensure all tests pass: `make test`
2. Run `make fmt` to format your code
3. Update documentation as needed
4. Add/update tests for your changes

### PR Checklist
Your PR should include:
- [ ] Clear description of changes and purpose
- [ ] Code follows existing design conventions
- [ ] No exponential time complexity
- [ ] No n+1 network request patterns
- [ ] Updated relevant documentation
- [ ] Comprehensive test coverage
- [ ] All tests pass

### PR Description
Please include:
- **Purpose**: What problem does this solve or what feature does it add?
- **Changes**: What specific changes were made?
- **Compromises**: Any trade-offs or decisions made during implementation
- **Risks**: Any potential risks introduced by the changes

## License


By contributing to this project, you agree that your contributions will be licensed under the Apache License 2.0.
