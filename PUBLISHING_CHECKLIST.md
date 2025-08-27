# Publishing Checklist for AI Provider Wrapper

Use this checklist to ensure your package is ready for publishing to the Go package registry.

## ✅ Pre-Publishing Checklist

### 📁 Package Structure
- [ ] Remove `main.go` file (library packages shouldn't have main functions)
- [ ] Ensure `go.mod` has correct module path
- [ ] All import paths use the correct module name
- [ ] Package follows Go project layout conventions

### 📄 Required Files
- [ ] `README.md` with installation and usage instructions
- [ ] `LICENSE` file (MIT, Apache 2.0, etc.)
- [ ] `go.mod` with appropriate Go version
- [ ] `.gitignore` with Go-specific ignores
- [ ] `CHANGELOG.md` documenting versions and changes

### 🧪 Code Quality
- [ ] All tests pass: `go test ./...`
- [ ] No issues with `go vet ./...`
- [ ] Package builds successfully: `go build ./...`
- [ ] Dependencies are clean: `go mod tidy && go mod verify`
- [ ] No unused dependencies

### 📚 Documentation
- [ ] All public functions have godoc comments
- [ ] README includes:
  - [ ] Installation instructions (`go get ...`)
  - [ ] Quick start example
  - [ ] API overview
  - [ ] Link to full documentation
- [ ] Examples are working and up-to-date
- [ ] Architecture documentation exists

### 🔧 Git Repository
- [ ] Repository is public on GitHub
- [ ] All changes are committed
- [ ] Working directory is clean (`git status`)
- [ ] Remote origin is set correctly

### 🏷️ Versioning
- [ ] Choose appropriate version number (semantic versioning)
- [ ] Update CHANGELOG.md with new version
- [ ] Create git tag: `git tag v1.0.0`
- [ ] Push tag: `git push origin v1.0.0`

## 🚀 Publishing Steps

### 1. Final Preparation
```bash
# Run the preparation script
./scripts/prepare-publish.sh

# Or manually run these commands:
go mod tidy
go test ./...
go vet ./...
go build ./...
```

### 2. Commit and Tag
```bash
# Commit final changes
git add .
git commit -m "Prepare for v1.0.0 release"

# Create and push tag
git tag v1.0.0
git push origin main
git push origin v1.0.0
```

### 3. Verify Publishing
```bash
# Check if package is available (may take a few minutes)
go list -m github.com/your-username/ai-provider-wrapper@v1.0.0

# Visit pkg.go.dev page
open https://pkg.go.dev/github.com/your-username/ai-provider-wrapper
```

### 4. Create GitHub Release (Optional but Recommended)
- Go to GitHub repository
- Click "Releases" → "Create a new release"
- Select tag `v1.0.0`
- Add release notes from CHANGELOG.md
- Publish release

## 📋 Post-Publishing Tasks

### 🔍 Verification
- [ ] Package appears on pkg.go.dev
- [ ] Documentation renders correctly
- [ ] Installation works: `go get github.com/your-username/ai-provider-wrapper`
- [ ] Examples in README work with published package

### 📈 Promotion
- [ ] Add badges to README:
  - [ ] Go Reference badge
  - [ ] Go Report Card badge
  - [ ] License badge
  - [ ] Build status badge
- [ ] Share on relevant communities (Reddit, Discord, etc.)
- [ ] Consider writing a blog post about the package

### 🔄 Maintenance Setup
- [ ] Set up GitHub Actions for CI/CD
- [ ] Configure automated security scanning
- [ ] Set up issue templates
- [ ] Create contributing guidelines
- [ ] Set up automated dependency updates (Dependabot)

## 🛠️ Quick Commands Reference

```bash
# Test everything
go test ./...

# Check for issues
go vet ./...

# Clean dependencies
go mod tidy

# Verify module
go mod verify

# Build package
go build ./...

# Create tag
git tag v1.0.0

# Push tag
git push origin v1.0.0

# Check package availability
go list -m github.com/your-username/ai-provider-wrapper@latest

# Install package (test)
go get github.com/your-username/ai-provider-wrapper
```

## 🚨 Common Issues and Solutions

### Package Not Appearing on pkg.go.dev
- Ensure repository is public
- Check that go.mod is valid
- Wait up to 30 minutes for indexing
- Try accessing the URL directly

### Import Path Issues
- Verify all internal imports use correct module path
- Run `go mod tidy` to clean up
- Check that module path matches repository URL

### Version Not Updating
- Ensure git tag is pushed: `git push origin v1.0.0`
- Clear module cache: `go clean -modcache`
- Use specific version: `go get github.com/your-username/ai-provider-wrapper@v1.0.0`

### Tests Failing
- Run tests locally: `go test -v ./...`
- Check for environment-specific issues
- Ensure all dependencies are available

## 📞 Getting Help

If you encounter issues:

1. Check the [Go Modules Reference](https://golang.org/ref/mod)
2. Visit [pkg.go.dev documentation](https://pkg.go.dev/about)
3. Ask on [Go Community](https://golang.org/help)
4. Check [GitHub Issues](https://github.com/golang/go/issues) for similar problems

---

**Ready to publish?** Run through this checklist and use the preparation script to ensure everything is ready! 🚀