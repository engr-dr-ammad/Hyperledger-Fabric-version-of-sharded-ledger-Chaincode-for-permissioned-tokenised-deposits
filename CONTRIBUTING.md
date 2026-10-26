# Contributing to Hyperledger Fabric Sharded Ledger Chaincode

Thank you for your interest in contributing! This document provides guidelines and instructions for contributing to this project.

## Code of Conduct

Please read and adhere to our [Code of Conduct](CODE_OF_CONDUCT.md).

## Getting Started

### Prerequisites
- Go 1.20+
- Hyperledger Fabric 2.5.0+
- Docker & Docker Compose
- Git

### Setup Development Environment

```bash
# Clone the repository
git clone https://github.com/engr-dr-ammad/hyperledger-fabric-sharded-ledger.git
cd hyperledger-fabric-sharded-ledger

# Install dependencies
cd chaincode
go mod download

# Run tests
go test -v ./...
```

## Development Workflow

### 1. Create a Feature Branch

```bash
git checkout -b feature/your-feature-name
# or for bugs:
git checkout -b fix/your-bug-fix
```

Use descriptive branch names:
- `feature/deposit-transfer-history` ✓
- `fix/permission-check-bug` ✓
- `docs/add-api-examples` ✓
- `refactor/sharding-logic` ✓

### 2. Make Your Changes

#### Code Style
- Follow standard Go conventions
- Use `gofmt` for formatting: `go fmt ./...`
- Use `golangci-lint` for linting: `golangci-lint run`
- Run `go vet` to check for issues: `go vet ./...`

#### Commit Messages
Write clear, descriptive commit messages:

```
Short summary (50 chars or less)

More detailed explanation if necessary. Wrap at 72 characters.
Explain what and why, not how.

Fixes #123
Co-Authored-By: Name <email@example.com>
```

#### Code Quality
- Write tests for new functionality
- Ensure all tests pass: `go test -v ./...`
- Maintain or improve code coverage
- Add comments for exported functions and complex logic
- Keep functions focused and small

### 3. Testing

#### Unit Tests
```bash
cd chaincode
go test -v -race -cover ./...
```

#### Integration Tests
```bash
# Start the network
docker-compose up -d

# Run integration tests
./scripts/setup-network.sh
./scripts/deploy-chaincode.sh
./scripts/init-ledger.sh
```

#### Coverage
```bash
cd chaincode
go test -v -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### 4. Documentation

- Update README.md if adding/removing features
- Update relevant docs in `/docs` folder
- Add inline code comments for complex logic
- Include examples for new API functions

### 5. Submit a Pull Request

#### Before Submitting
- [ ] Branch is up to date with master: `git rebase origin/master`
- [ ] All tests pass locally
- [ ] Code is formatted: `go fmt ./...`
- [ ] Code passes lint checks: `golangci-lint run`
- [ ] No new warnings from `go vet`
- [ ] Test coverage is maintained/improved
- [ ] Documentation is updated

#### Creating the PR
1. Push your branch to your fork
2. Create a PR against the master branch
3. Fill out the PR template completely
4. Link any related issues
5. Request review from maintainers

#### PR Requirements
- Clear description of changes
- Reference to related issues (if any)
- Test results showing all tests pass
- Documentation updates (if applicable)
- Sign-off with: `Co-Authored-By: Your Name <your.email@example.com>`

## Code Review Process

### What to Expect
- At least one maintainer review
- Automated tests must pass
- Code style and quality checks
- Feedback may be requested
- Changes may be requested before approval

### Providing Feedback
When reviewing code:
- Be respectful and constructive
- Explain the reasoning behind suggestions
- Acknowledge good work and improvements
- Ask questions rather than making demands

## Issues and Bug Reports

### Finding Issues to Work On
- Look for issues labeled `good first issue`
- Check labels like `help wanted`, `feature`, `bug`
- Discuss larger changes before starting work

### Reporting Bugs
Use the [bug report template](.github/ISSUE_TEMPLATE/bug_report.md):
- Clear description of the issue
- Steps to reproduce
- Expected vs actual behavior
- Environment details
- Error messages/logs

### Requesting Features
Use the [feature request template](.github/ISSUE_TEMPLATE/feature_request.md):
- Clear description of the feature
- Use case and motivation
- Proposed solution
- Alternatives considered

## Documentation

### Update Documentation When
- Adding new functions or features
- Changing API behavior
- Fixing bugs that affect usage
- Improving existing documentation

### Documentation Standards
- Write in clear, concise English
- Include code examples
- Update table of contents if adding sections
- Keep formatting consistent

## Security

### Reporting Security Issues
**Do NOT create a public issue for security vulnerabilities.**

Instead:
1. Email security details to the maintainers
2. Include proof of concept if possible
3. Allow time for a fix before public disclosure
4. Follow responsible disclosure practices

### Security Best Practices
- Don't commit secrets, keys, or credentials
- Use `.gitignore` for sensitive files
- Review dependencies for vulnerabilities
- Follow OWASP guidelines

## Version Control

### Branching Strategy
```
master (production-ready)
  ↑
  ├─ feature/xyz
  ├─ fix/abc
  └─ docs/def
```

### Commit Best Practices
- Make small, logical commits
- One feature/fix per commit when possible
- Write descriptive commit messages
- Don't commit unrelated changes together

### Keeping Branches Updated
```bash
# Fetch latest changes
git fetch origin

# Rebase on master
git rebase origin/master

# Or merge if you prefer
git merge origin/master
```

## Release Process

### Version Numbers
We use [Semantic Versioning](https://semver.org/):
- MAJOR.MINOR.PATCH (e.g., 1.2.3)
- MAJOR: Breaking changes
- MINOR: New features
- PATCH: Bug fixes

### Creating a Release
1. Update version in documentation
2. Update CHANGELOG.md
3. Create git tag: `git tag v1.2.3`
4. Push tag: `git push origin v1.2.3`
5. Create GitHub Release with notes

## Community

### Communication Channels
- GitHub Issues: For bugs, features, questions
- GitHub Discussions: For conversations and ideas
- Email: For security issues

### Getting Help
- Check existing issues and discussions
- Review documentation
- Ask in issues with clear context
- Be patient and respectful

## Code Examples

### Adding a New Chaincode Function

```go
// NewFunction performs a specific action
// 
// Parameters:
// - ctx: Transaction context
// - param1: Description of param1
// - param2: Description of param2
// 
// Returns: Error if operation fails
func (dc *DepositChaincode) NewFunction(
    ctx contractapi.TransactionContextInterface,
    param1 string,
    param2 string,
) error {
    // Validate permissions
    if err := dc.permissions.RequireRole(ctx, "admin"); err != nil {
        return fmt.Errorf("unauthorized: %v", err)
    }

    // Validate inputs
    if param1 == "" {
        return fmt.Errorf("param1 cannot be empty")
    }

    // Perform operation
    // ... implementation ...

    // Emit event
    return dc.ledgerManager.EmitEvent(ctx, "eventName", data)
}
```

### Writing Tests

```go
func TestNewFunction(t *testing.T) {
    cc := NewDepositChaincode()
    stub := NewMockStub()
    ctx := &MockTransactionContext{
        stub: stub,
        clientID: &MockClientID{
            mspID: "TestMSP",
            cn:    "test-user",
        },
    }

    // Test case 1: Valid operation
    err := cc.NewFunction(ctx, "valid", "params")
    assert.NoError(t, err)

    // Test case 2: Invalid input
    err = cc.NewFunction(ctx, "", "invalid")
    assert.Error(t, err)
}
```

## Questions?

- Create an issue with the `question` label
- Check existing documentation
- Ask in GitHub Discussions
- Email the maintainers

## Recognition

Contributors are recognized in:
- CONTRIBUTORS.md
- Release notes
- Commit history

Thank you for contributing! 🎉
