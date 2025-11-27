# Pilgrim

> [!WARNING]
> This is purely a toy project for small back end projects I'm building. Please don't use this for anything serious.

Skipping the ORM: _Pilgrim_ is a Simple migration CLI for managing database schema changes.

## Databases Supported
* PostgreSQL

## Setup

Install Go dependencies and either build the binary or run directly with `go run`.

```bash
go mod download

# Option 1: Build the binary
go build -o pilgrim cmd/pilgrim/main.go

# Option 2: Run directly
go run cmd/pilgrim/main.go
```

See the [configuration documentation](docs/PILGRIM-CONFIG.md) for details on setting up migration directories and a database connection.

### Running Tests

Unit tests can be run with:

```bash
sh scripts/unit-tests.sh
```

End-to-end tests can be found under the `e2e` directory. They can be run with:

```bash
cd e2e && python tests.py
```

## FAQ

### Why the name Pilgrim?
I'm lazy; I looked up synonyms for "migrate" and found "pilgrim".

### Why not use an existing migration tool?
For small projects where an ORM is not needed, existing migration tools can feel like overkill. Pilgrim aims to be a lightweight alternative for such scenarios.

### How do I contribute?
Feel free to open issues or submit pull requests on GitHub. Contributions are welcome!

### Does Pilgrim automatically generate SQL when creating migration files?
No, Pilgrim does not generate SQL automatically. Users are expected to write their own SQL in the migration files. Pilgrim focuses on managing and applying these migrations rather than generating them.

### What platofrms does Pilgrim support?
Pilgrim is currently only operational on AMD64 Linux systems. Support for additional platforms may be added in the future based on demand.

