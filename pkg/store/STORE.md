# Store

`dployr` uses an embedded SQLite database for persistence, with migrations and queries structured for reliability and performance.

---

## Database

- **Location:** `~/.dployr/data.db`  
- **Driver:** [`modernc.org/sqlite`](https://pkg.go.dev/modernc.org/sqlite) (pure Go, no CGO)

---

## Core Tables

- **`users`** – user accounts and role-based permissions (`owner`, `admin`, `developer`, `viewer`)
- **`deployments`** – deployment records containing JSON configuration and status tracking
- **`services`** – active service instance linked to deployments
- **`schema_migrations`** – tracks applied migrations for database evolution

---

## Migration System

Migrations are embedded directly into the binary and applied automatically at startup.

- **Location:** [`internal/db/migrations/`](../../internal/db/migrations/)  
- **Format:** Sequential SQL files (`000_init.sql`, `001_feature.sql`, …)  
- **Execution:** Each migration runs within an atomic transaction and rolls back on failure  
- **Tracking:** Applied migrations are recorded in the `schema_migrations` table  

This design ensures consistent schema evolution across environments without requiring manual setup.

---

## Query Implementation

Store interfaces and their implementations follow a clear repository-style pattern.

- **Interfaces:** [`pkg/store/`](.)  
- **Implementations:** [`internal/store/`](../../internal/store/)  
- **Pattern:** Interface segregation with context-aware prepared statements  

---

## Implementation Notes

- Complex configuration objects in `deployments.config` are serialized using JSON  
- Foreign key relationships maintain links between `services` and `deployments`  
- List operations support pagination via `LIMIT/OFFSET`  
- All queries are context-aware, supporting cancellation and timeout behavior  

---

### Example Flow

1. On startup, the store initializes and applies pending migrations.  
2. Prepared statements are created for core CRUD operations.  
3. Each operation runs within a request-scoped context, allowing cancellation or deadline enforcement.  
4. Results are returned as typed Go structs, with JSON fields automatically marshaled/unmarshaled.  

---

## Summary

The store package provides a clean, dependable abstraction over SQLite, balancing simplicity with performance.  
By embedding migrations and using prepared statements, `dployr` ensures consistent behavior across environments without external dependencies.

---
