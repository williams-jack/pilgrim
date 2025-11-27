# Pilgrim Configuration Guide

This guide describes how to create a configuration file for Pilgrim, specifying database connection details and migration directories. The configuration is written in JSON.

---

## Top-Level Structure

Your configuration file should be a JSON object with the following fields:

| Field             | Type                | Description                                               |
|-------------------|---------------------|-----------------------------------------------------------|
| dbType            | string              | The type of database (e.g., `"postgres"`).                |
| downDir           | string              | Directory containing "down" migration scripts.             |
| upDir             | string              | Directory containing "up" migration scripts.               |
| connectionString  | string (optional)   | Direct database connection string (overrides dbConfig).    |
| dbConfig          | object              | Database-specific configuration (see below).               |

NOTE: If the directories specified in `downDir` and `upDir` do not exist yet, Pilgrim can create them automatically given the `pilgrim init` command by passing the `-d` flag.

---

## Example Configuration (PostgreSQL)

```json
{
  "dbType": "postgres",
  "downDir": "./migrations/down",
  "upDir": "./migrations/up",
  "dbConfig": {
    "host": "localhost",
    "port": 5432,
    "user": "pilgrim_user",
    "password": "your_password",
    "dbName": "pilgrim_db",
    "params": {
      "sslmode": "disable"
    }
  }
}
```

### Field Details

#### dbType

- **Required**
- Supported value: `"postgres"`

#### downDir / upDir

- **Required**
- Path to migration directories.

#### connectionString

- **Optional**
- If provided, this string is used directly to connect to the database.
- If omitted, Pilgrim will construct the connection string from `dbConfig`.

#### dbConfig (for PostgreSQL)

| Field     | Type              | Description                                 |
|-----------|-------------------|---------------------------------------------|
| host      | string            | Database hostname or IP address             |
| port      | integer           | Database port number                        |
| user      | string            | Username for authentication                 |
| password  | string            | Password for authentication                 |
| dbName    | string            | Name of the database to connect to          |
| params    | object (key-value)| Additional connection parameters (optional) |

Example `params`:
```json
"params": {
  "sslmode": "disable",
  "application_name": "pilgrim"
}
```

---

## Notes

* Only `"postgres"` is currently supported for `dbType`.
* If both `connectionString` and `dbConfig` are provided, `connectionString` takes precedence.
* All fields in `dbConfig` are optional, but missing fields may result in an invalid connection string.

---

## Usage

It's recommended to save your configuration file as `pilgrim.config.json` in your project root. You can then use this configuration file with Pilgrim commands to manage your database migrations.

## Using Environment Variables in Configuration

Pilgrim supports the use of environment variables in your `pilgrim.config.json` file. You can reference environment variables in any string field by using the `${VAR_NAME}` syntax. At runtime, Pilgrim will automatically replace these placeholders with the corresponding environment variable values. If the environment variable is not set, Pilgrim will replace it with an empty string.

**Example:**
```json
{
  "dbType": "postgres",
  "dbConfig": {
    "host": "${PG_HOST}",
    "port": "${PG_PORT}",
    "user": "${PG_USER}",
    "password": "${PG_PASSWORD}",
    "dbName": "${PG_DBNAME}"
  }
}

In this example, the values for `host`, `port`, `user`, `password`, and `dbName` will be taken from the respective environment variables when Pilgrim runs.
