# HubInvestmentsServer

## 🚀 Quick Coverage Commands

**Generate coverage for the ENTIRE project and open HTML report:**

```bash
make coverage-open
```

**Alternative commands for the same result:**
```bash
# Using bash script (with colored output)
./scripts/coverage.sh open

# Manual step-by-step
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
open coverage.html  # macOS
```

**Other useful coverage commands:**
```bash
make coverage-summary          # Show detailed coverage summary in terminal
make coverage                  # Show basic coverage percentages
make check                     # Run format + lint + tests + coverage summary
```

---

## 📊 Scripts Documentation

For detailed information about all available scripts and commands, see [scripts/README.md](scripts/README.md).

## 🎯 Development Workflow

1. **Quick coverage check**: `make coverage-open`
2. **Before committing**: `make check` 
3. **While writing tests**: `./scripts/test.sh watch`
