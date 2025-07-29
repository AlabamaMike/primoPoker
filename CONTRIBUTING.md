## Contributing to PrimoPoker

Thank you for your interest in contributing to PrimoPoker! This document provides guidelines and information for contributors.

### Getting Started

1. **Fork the repository** on GitHub
2. **Clone your fork** locally:
   ```bash
   git clone https://github.com/YOUR_USERNAME/primoPoker.git
   cd primoPoker
   ```
3. **Create a feature branch:**
   ```bash
   git checkout -b feature/your-feature-name
   ```

### Development Setup

1. **Install dependencies:**
   ```bash
   go mod download
   ```

2. **Copy environment variables:**
   ```bash
   cp .env.example .env
   # Edit .env with your local configuration
   ```

3. **Run tests:**
   ```bash
   go test ./tests/...
   ```

4. **Start the development server:**
   ```bash
   go run cmd/server/main.go
   ```

### Coding Standards

- **Follow Go conventions** and use `gofmt` for formatting
- **Write tests** for new functionality
- **Add documentation** for public functions and packages
- **Keep functions small** and focused on a single responsibility
- **Use meaningful variable names** and avoid abbreviations
- **Comment complex logic** and business rules

### Code Organization

- **`cmd/`** - Application entry points
- **`internal/`** - Private application code
- **`pkg/`** - Library code that can be used by other applications
- **`tests/`** - Test files
- **`.github/`** - GitHub-specific configuration

### Testing Guidelines

- **Unit tests** should cover core business logic
- **Integration tests** should test component interactions
- **Test files** should use the `_test.go` suffix
- **Use table-driven tests** for testing multiple scenarios
- **Mock external dependencies** in tests

### Commit Message Format

Use conventional commit format:

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
- `feat(game): add all-in betting functionality`
- `fix(websocket): resolve connection timeout issue`
- `docs(readme): update installation instructions`

### Pull Request Process

1. **Ensure all tests pass:**
   ```bash
   go test ./...
   ```

2. **Run code formatting:**
   ```bash
   gofmt -w .
   ```

3. **Update documentation** if needed

4. **Create a pull request** with:
   - Clear title and description
   - Reference to related issues
   - List of changes made
   - Testing instructions

5. **Wait for review** and address feedback

### Reporting Issues

When reporting bugs:

- **Use a clear title** describing the issue
- **Provide steps to reproduce** the bug
- **Include error messages** and logs
- **Specify your environment** (Go version, OS, etc.)
- **Add labels** to categorize the issue

### Feature Requests

When requesting features:

- **Explain the use case** and problem it solves
- **Provide examples** of how it would work
- **Consider implementation complexity**
- **Discuss alternatives** if applicable

### Security Issues

For security vulnerabilities:

- **DO NOT** create public issues
- **Email directly** to the maintainers
- **Provide detailed information** about the vulnerability
- **Wait for response** before public disclosure

### Questions and Support

- **Check existing issues** and documentation first
- **Use GitHub Discussions** for general questions
- **Join our community** channels (if available)
- **Be respectful** and patient with responses

### License

By contributing to this project, you agree that your contributions will be licensed under the same license as the project.

Thank you for contributing to PrimoPoker! 🎮
