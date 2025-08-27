# Publishing Guide: AI Provider Wrapper Go Package

This guide walks you through publishing the AI Provider Wrapper package to the Go package registry and making it available for public use.

## Prerequisites

Before publishing, ensure you have:

1. **Go 1.19+** installed
2. **Git** configured with your GitHub credentials
3. **GitHub repository** created and accessible
4. **Package properly structured** as a Go module

## Pre-Publishing Checklist

### 1. Fix Package Structure Issues

#### Remove main.go (Library packages shouldn't have main functions)
```bash
rm main.go
```

#### Ensure go.mod is properly configured
Your `go.mod` should look like:
```go
module github.com/your-username/ai-provider-wrapper

go 1.21
```

#### Update import paths in all files
Replace all internal imports from:
```go
"github.com/ai-provider-wrapper/ai-provider-wrapper/..."
```
To:
```go
"github.com/your-username/ai-provider-wrapper/..."
```

### 2. Create Essential Files

#### Create LICENSE file
```bash
# Choose an appropriate license (MIT is common for open source)
touch LICENSE
```

Example MIT License content:
```
MIT License

Copyright (c) 2024 Your Name

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```

#### Create .gitignore
```bash
# .gitignore
*.log
*.tmp
.env
.DS_Store
vendor/
dist/
coverage.out
*.test
```

#### Create CHANGELOG.md
```markdown
# Changelog

All notable changes to this project will be documented in this file.

## [v1.0.0] - 2024-01-XX

### Added
- Initial release of AI Provider Wrapper
- Support for OpenAI GPT models
- Support for Anthropic Claude models
- Unified interface for multiple AI providers
- Comprehensive error handling and retry logic
- Parameter validation and normalization
- Environment variable configuration support
- Extensive documentation and examples

### Features
- Text completion support
- Chat completion support
- Provider-agnostic parameter mapping
- Automatic retry with exponential backoff
- Rate limiting handling
- Token usage tracking
- Multiple configuration options
```

### 3. Validate Package Structure

Your final package structure should look like:
```
ai-provider-wrapper/
├── README.md
├── LICENSE
├── CHANGELOG.md
├── go.mod
├── go.sum (generated)
├── .gitignore
├── client.go
├── config.go
├── errors.go
├── interfaces.go
├── types.go
├── adapters/
│   ├── adapter.go
│   ├── openai/
│   │   ├── adapter.go
│   │   └── adapter_test.go
│   ├── anthropic/
│   │   ├── adapter.go
│   │   └── adapter_test.go
│   └── google/
│       └── adapter.go
├── internal/
│   ├── http/
│   │   └── client.go
│   └── utils/
│       ├── validation.go
│       └── validation_test.go
├── types/
│   └── types.go
├── docs/
│   ├── architecture-diagram.md
│   ├── providers.md
│   └── troubleshooting.md
└── examples/
    ├── README.md
    ├── basic/
    ├── chat/
    ├── config/
    ├── errors/
    ├── parameters/
    └── provider-switching/
```

## Publishing Steps

### Step 1: Prepare Your GitHub Repository

1. **Create GitHub Repository**
   ```bash
   # If not already created
   gh repo create your-username/ai-provider-wrapper --public
   ```

2. **Initialize Git (if not done)**
   ```bash
   git init
   git add .
   git commit -m "Initial commit: AI Provider Wrapper v1.0.0"
   git branch -M main
   git remote add origin https://github.com/your-username/ai-provider-wrapper.git
   git push -u origin main
   ```

### Step 2: Test Your Package

1. **Run all tests**
   ```bash
   go test ./...
   ```

2. **Check for issues**
   ```bash
   go vet ./...
   go mod tidy
   go mod verify
   ```

3. **Test package installation**
   ```bash
   # In a separate directory
   mkdir test-install
   cd test-install
   go mod init test
   go get github.com/your-username/ai-provider-wrapper
   ```

### Step 3: Create and Push Version Tags

1. **Create a version tag**
   ```bash
   git tag v1.0.0
   git push origin v1.0.0
   ```

2. **Verify tag creation**
   ```bash
   git tag -l
   ```

### Step 4: Publish to pkg.go.dev

The Go package registry (pkg.go.dev) automatically indexes public GitHub repositories. Once you push your tagged version:

1. **Trigger indexing** by visiting:
   ```
   https://pkg.go.dev/github.com/your-username/ai-provider-wrapper
   ```

2. **Or use the Go proxy directly**:
   ```bash
   GOPROXY=proxy.golang.org go list -m github.com/your-username/ai-provider-wrapper@latest
   ```

### Step 5: Create GitHub Release

1. **Create release on GitHub**
   ```bash
   gh release create v1.0.0 --title "AI Provider Wrapper v1.0.0" --notes-file CHANGELOG.md
   ```

2. **Or create manually**:
   - Go to your GitHub repository
   - Click "Releases" → "Create a new release"
   - Choose tag `v1.0.0`
   - Add release title and description
   - Publish release

## Post-Publishing Tasks

### 1. Update Documentation

Ensure your README.md includes:
- Installation instructions
- Quick start guide
- API documentation links
- Examples
- Contributing guidelines

### 2. Set Up Continuous Integration

Create `.github/workflows/ci.yml`:
```yaml
name: CI

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]

jobs:
  test:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        go-version: [1.19, 1.20, 1.21]

    steps:
    - uses: actions/checkout@v3

    - name: Set up Go
      uses: actions/setup-go@v3
      with:
        go-version: ${{ matrix.go-version }}

    - name: Build
      run: go build -v ./...

    - name: Test
      run: go test -v ./...

    - name: Vet
      run: go vet ./...
```

### 3. Add Package Badges

Add to your README.md:
```markdown
[![Go Reference](https://pkg.go.dev/badge/github.com/your-username/ai-provider-wrapper.svg)](https://pkg.go.dev/github.com/your-username/ai-provider-wrapper)
[![Go Report Card](https://goreportcard.com/badge/github.com/your-username/ai-provider-wrapper)](https://goreportcard.com/report/github.com/your-username/ai-provider-wrapper)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
```

## Usage After Publishing

Once published, users can install your package:

```bash
go get github.com/your-username/ai-provider-wrapper
```

And use it in their code:
```go
import "github.com/your-username/ai-provider-wrapper"

func main() {
    client, err := aiprovider.NewClient(aiprovider.ProviderOpenAI, aiprovider.Config{
        APIKey: "your-api-key",
    })
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()
    
    // Use the client...
}
```

## Version Management

### Semantic Versioning

Follow semantic versioning (semver):
- **v1.0.0**: Initial release
- **v1.0.1**: Bug fixes
- **v1.1.0**: New features (backward compatible)
- **v2.0.0**: Breaking changes

### Creating New Versions

```bash
# For bug fixes
git tag v1.0.1
git push origin v1.0.1

# For new features
git tag v1.1.0
git push origin v1.1.0

# For breaking changes
git tag v2.0.0
git push origin v2.0.0
```

## Troubleshooting

### Common Issues

1. **Package not appearing on pkg.go.dev**
   - Ensure repository is public
   - Check that go.mod is valid
   - Try accessing the package URL directly
   - Wait up to 30 minutes for indexing

2. **Import path issues**
   - Ensure all internal imports use the correct module path
   - Run `go mod tidy` to clean up dependencies

3. **Version not updating**
   - Ensure you've pushed the git tag
   - Clear Go module cache: `go clean -modcache`
   - Use specific version: `go get github.com/your-username/ai-provider-wrapper@v1.0.0`

### Validation Commands

```bash
# Check module validity
go mod verify

# List module versions
go list -m -versions github.com/your-username/ai-provider-wrapper

# Check what version is being used
go list -m github.com/your-username/ai-provider-wrapper

# Force update to latest
go get -u github.com/your-username/ai-provider-wrapper
```

## Best Practices

1. **Documentation**: Maintain comprehensive documentation
2. **Testing**: Ensure good test coverage
3. **Versioning**: Follow semantic versioning strictly
4. **Backwards Compatibility**: Avoid breaking changes in minor versions
5. **Security**: Regularly update dependencies
6. **Community**: Respond to issues and pull requests promptly

## Resources

- [Go Modules Reference](https://golang.org/ref/mod)
- [pkg.go.dev Documentation](https://pkg.go.dev/about)
- [Semantic Versioning](https://semver.org/)
- [Go Module Publishing Best Practices](https://blog.golang.org/publishing-go-modules)

---

**Next Steps**: Follow the checklist above, fix any issues, and publish your first version!