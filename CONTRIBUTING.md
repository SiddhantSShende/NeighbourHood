# Contributing to NeighbourHood

Thank you for considering contributing to NeighbourHood! This document outlines the process and guidelines for contributing.

## 🎯 How to Contribute

### Reporting Bugs

1. Check if the bug has already been reported in [Issues](../../issues)
2. If not, create a new issue with:
   - Clear, descriptive title
   - Steps to reproduce
   - Expected vs actual behavior
   - Environment details (OS, Go version, etc.)
   - Screenshots if applicable

### Suggesting Features

1. Check existing feature requests in [Issues](../../issues)
2. Create a new issue with:
   - Clear description of the feature
   - Use cases and benefits
   - Potential implementation approach

### Pull Requests

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes following our coding standards
4. Write or update tests as needed
5. Update documentation as needed
6. Commit your changes (`git commit -m 'Add amazing feature'`)
7. Push to the branch (`git push origin feature/amazing-feature`)
8. Open a Pull Request

## 📝 Coding Standards

### Go Code Style

- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Run `go fmt` before committing
- Use `goimports` for import formatting
- Run `golangci-lint run` to check for issues

### Clean Code Principles

1. **Single Responsibility**: Each function/struct should do one thing
2. **DRY (Don't Repeat Yourself)**: Avoid code duplication
3. **Descriptive Names**: Use clear, self-documenting names
4. **Small Functions**: Keep functions focused and concise
5. **Error Handling**: Always handle errors explicitly

### Project Structure

```
internal/
├── package/
│   └── package.go      # Main package code
│   └── package_test.go # Tests for package
```

- Keep packages focused on a single domain
- Use dependency injection
- Avoid circular dependencies

### Code Example

```go
// Good: Clear, focused, well-documented
package integrations

// SendMessage sends a message to a Slack channel
func (p *SlackProvider) SendMessage(ctx context.Context, channel, text string) error {
    if channel == "" {
        return errors.New("channel cannot be empty")
    }
    if text == "" {
        return errors.New("text cannot be empty")
    }
    
    // Implementation
    return nil
}

// Bad: Unclear, doing too much
func (p *SlackProvider) Do(ctx context.Context, stuff map[string]interface{}) interface{} {
    // Multiple responsibilities, unclear purpose
}
```

## 🧪 Testing

### Writing Tests

- Write unit tests for all new code
- Use table-driven tests where appropriate
- Mock external dependencies
- Aim for >80% code coverage

### Example Test

```go
func TestSlackProvider_SendMessage(t *testing.T) {
    tests := []struct {
        name    string
        channel string
        text    string
        wantErr bool
    }{
        {
            name:    "valid message",
            channel: "#general",
            text:    "Hello",
            wantErr: false,
        },
        {
            name:    "empty channel",
            channel: "",
            text:    "Hello",
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            p := NewSlackProvider("id", "secret", "url")
            err := p.SendMessage(context.Background(), tt.channel, tt.text)
            if (err != nil) != tt.wantErr {
                t.Errorf("SendMessage() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run specific package tests
go test ./internal/integrations/...
```

## 📦 Adding New Integrations

To add a new integration provider:

1. **Define the provider struct** in `internal/integrations/integrations.go`:
```go
type NewProvider struct {
    ClientID     string
    ClientSecret string
    RedirectURL  string
}
```

2. **Implement the Provider interface**:
```go
func (p *NewProvider) Name() string { /* ... */ }
func (p *NewProvider) GetAuthURL(state string) string { /* ... */ }
func (p *NewProvider) ExchangeCode(ctx context.Context, code string) (*Token, error) { /* ... */ }
func (p *NewProvider) Execute(ctx context.Context, token *Token, action string, payload map[string]interface{}) (interface{}, error) { /* ... */ }
```

3. **Add configuration** in `internal/config/config.go`

4. **Register the provider** in `cmdapi/main.go`

5. **Write tests** for the new provider

6. **Update documentation**:
   - README.md
   - API_DOCUMENTATION.md

## 🔄 Git Workflow

### Branching Strategy

- `main`: Production-ready code
- `develop`: Development branch
- `feature/*`: New features
- `bugfix/*`: Bug fixes
- `hotfix/*`: Urgent production fixes

### Commit Messages

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Maintenance tasks

**Examples:**
```
feat(integrations): add Gmail provider
fix(workflow): handle empty token map
docs(api): update workflow execution examples
```

## 🔍 Code Review Process

### For Contributors

- Ensure all tests pass
- Update documentation
- Keep PRs focused and small
- Respond to review feedback promptly

### For Reviewers

- Check code quality and adherence to standards
- Verify tests are comprehensive
- Ensure documentation is updated
- Be constructive and respectful

## 🏗️ Architecture Decisions

### When to Add a New Package

Create a new package when:
- You have a distinct domain concern
- The package would be >500 lines
- You need to hide internal implementation details

### When to Refactor

Refactor when:
- Code is duplicated in 3+ places
- Function is >50 lines
- Cyclomatic complexity is too high
- Tests are difficult to write

## 📚 Documentation

### Code Documentation

- Document all exported functions, types, and constants
- Use complete sentences
- Include examples for complex functions

```go
// GetAuthURL returns the OAuth 2.0 authorization URL for the provider.
// The state parameter should be a random string to prevent CSRF attacks.
//
// Example:
//   url := provider.GetAuthURL("random-state-123")
//   // Returns: "https://provider.com/oauth?client_id=...&state=random-state-123"
func (p *Provider) GetAuthURL(state string) string {
    // Implementation
}
```

### Project Documentation

- Update README.md for major changes
- Update API_DOCUMENTATION.md for API changes
- Add examples for new features

## ✅ Pre-Commit Checklist

Before submitting a PR, ensure:

- [ ] Code follows project style guidelines
- [ ] All tests pass (`make test`)
- [ ] Code is formatted (`make format`)
- [ ] Linter passes (`make lint`)
- [ ] Documentation is updated
- [ ] Commit messages follow convention
- [ ] No sensitive data in commits

## 🆘 Getting Help

- Open an issue for questions
- Join our community chat (if available)
- Check existing documentation and issues first

## 📄 License

By contributing, you agree that your contributions will be licensed under the project's license.

---

Thank you for contributing to NeighbourHood! 🎉
