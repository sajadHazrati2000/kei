# Contributing to Kei

Thanks for your interest in contributing. Kei is in active early development — the best way to help right now is to try deploying it and report issues.

## Before You Start

- Read [PRODUCT.md](PRODUCT.md) — understand the scope and what's intentionally out of scope
- Check open issues on GitHub before starting work
- For significant changes, open an issue first to discuss

## Local Development Setup

```bash
# Clone
git clone https://github.com/your-handle/kei.git
cd kei

# Backend
cd backend
go mod download
go run cmd/server/main.go

# Frontend
cd frontend
npm install
npm start
```

## Pull Request Checklist

- [ ] Code follows existing patterns (handler → service → repository)
- [ ] New endpoints have corresponding tests
- [ ] No new dependencies added without discussion
- [ ] Changes are within current phase scope
- [ ] Tested locally against Docker Compose setup

## Code Style

- Go: `gofmt` formatted, errors wrapped with context
- Angular: standalone components, signals for state, `inject()` over constructor
- Commits: conventional commits format (`feat:`, `fix:`, `docs:`)

## Reporting Issues

Use GitHub Issues. Include:
- Deployment method (Docker / local dev)
- Steps to reproduce
- Expected vs actual behavior
