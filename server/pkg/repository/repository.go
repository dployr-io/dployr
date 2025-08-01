package repository

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type QuerySet struct {
	getByID string
	create  string
	update  string
	delete  string
	upsert  string
}

type AppRepos struct {
	ProjectRepo      *ProjectRepo
	UserRepo         *UserRepo
	MagicTokenRepo   *MagicTokenRepo
	RefreshTokenRepo *RefreshTokenRepo
	EventRepo        *EventRepo
	DeploymentRepo   *DeploymentRepo
	LogRepo          *LogRepo
}

// Get struct fields using reflection
func getStructFields(entity interface{}) []string {
	t := reflect.TypeOf(entity)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	var fields []string
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if dbTag := field.Tag.Get("db"); dbTag != "" && dbTag != "-" {
			fields = append(fields, dbTag)
		}
	}
	return fields
}

// Filter out specific fields
func filterFields(fields []string, exclude string) []string {
	var result []string
	for _, field := range fields {
		if field != exclude {
			result = append(result, field)
		}
	}
	return result
}

// Build named placeholders for INSERT
func buildPlaceholders(fields []string) string {
	var placeholders []string
	for _, field := range fields {
		placeholders = append(placeholders, ":"+field)
	}
	return strings.Join(placeholders, ", ")
}

// Build UPDATE clause with named parameters
func buildUpdateClause(fields []string) string {
	var updates []string
	for _, field := range fields {
		if field != "created_at" { // Don't update created_at
			updates = append(updates, fmt.Sprintf("%s = :%s", field, field))
		}
	}
	return strings.Join(updates, ", ")
}

// Build optimized SQL queries based on struct fields
func buildQueries(tableName string, fields []string) QuerySet {
	// Remove ID from insert fields
	insertFields := filterFields(fields, "id")
	insertFields = filterFields(insertFields, "created_at")
	placeholders := buildPlaceholders(insertFields)

	return QuerySet{
		getByID: fmt.Sprintf("SELECT %s FROM %s WHERE id = :id LIMIT 1",
			strings.Join(fields, ", "), tableName),

		create: fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) RETURNING id",
			tableName, strings.Join(insertFields, ", "), placeholders),

		update: fmt.Sprintf("UPDATE %s SET %s WHERE id = :id",
			tableName, buildUpdateClause(insertFields)),

		delete: fmt.Sprintf("DELETE FROM %s WHERE id = :id", tableName),

		upsert: "", // Built dynamically in Upsert method
	}
}

// Check if error is primary key violation
func isPrimaryKeyViolation(err error) bool {
	if pqErr, ok := err.(*pq.Error); ok {
		return pqErr.Code == "23505" // unique_violation
	}
	return false
}

// Set ID field if it exists (for auto-generated IDs)
func setIDIfExists(entity interface{}, id int64) {
	v := reflect.ValueOf(entity)

	// Handle both pointer and value types
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return
		}
		v = v.Elem()
	} else {
		// For value types, we can't modify the original, so return early
		return
	}

	idField := v.FieldByName("ID")
	if idField.IsValid() && idField.CanSet() {
		if idField.Kind() == reflect.Int64 {
			idField.SetInt(id)
		} else if idField.Kind() == reflect.String {
			idField.SetString(fmt.Sprintf("%d", id))
		}
	}
}

type Repository[T any] struct {
	db        *sqlx.DB
	tableName string
	queries   QuerySet
}

type SQLRepository[T any] interface {
	Create(ctx context.Context, entity *T) error
	Update(ctx context.Context, entity T) error
	Upsert(ctx context.Context, entity T, conflictCols []string, updateCols []string) error
	Delete(ctx context.Context, id any) error
	GetByID(ctx context.Context, id any) error
}

// NewRepository creates a new generic repository with direct queries
func NewRepository[T any](db *sqlx.DB, tableName string) *Repository[T] {
	var entity T
	fields := getStructFields(entity)
	queries := buildQueries(tableName, fields)

	return &Repository[T]{
		db:        db,
		tableName: tableName,
		queries:   queries,
	}
}

func (r *Repository[T]) Create(ctx context.Context, entity *T) error {
	result, err := r.db.NamedExecContext(ctx, r.queries.create, entity)
	if err != nil {
		if isPrimaryKeyViolation(err) {
			return fmt.Errorf("record already exists: %w", err)
		}
		return fmt.Errorf("create failed: %w", err)
	}

	if id, err := result.LastInsertId(); err == nil && id > 0 {
		setIDIfExists(entity, id)
	}

	return nil
}

func (r *Repository[T]) GetByID(ctx context.Context, id any) (T, error) {
	var entity T

	rows, err := r.db.NamedQueryContext(ctx, r.queries.getByID, map[string]interface{}{
		"id": id,
	})
	if err != nil {
		return entity, fmt.Errorf("get by id failed: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return entity, fmt.Errorf("record not found")
	}

	err = rows.StructScan(&entity)
	if err != nil {
		return entity, fmt.Errorf("get by id scan failed: %w", err)
	}

	return entity, nil
}

func (r *Repository[T]) Update(ctx context.Context, entity T) error {
	result, err := r.db.NamedExecContext(ctx, r.queries.update, entity)
	if err != nil {
		return fmt.Errorf("update failed: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("record not found")
	}

	return nil
}

func (r *Repository[T]) Delete(ctx context.Context, id any) error {
	result, err := r.db.NamedExecContext(ctx, r.queries.delete, map[string]interface{}{
		"id": id,
	})
	if err != nil {
		return fmt.Errorf("delete failed: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("record not found")
	}

	return nil
}

// Build UPSERT query with ON CONFLICT
func (r *Repository[T]) BuildUpsertQuery(conflictCols []string, updateCols []string) string {
	var entity T
	fields := getStructFields(entity)

	var insertFields []string
	for _, field := range fields {
		if field != "id" && field != "created_at" {
			insertFields = append(insertFields, field)
		}
	}
	placeholders := buildPlaceholders(insertFields)

	conflictClause := strings.Join(conflictCols, ", ")

	var updates []string
	for _, col := range updateCols {
		updates = append(updates, fmt.Sprintf("%s = EXCLUDED.%s", col, col))
	}
	updateClause := strings.Join(updates, ", ")

	return fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) ON CONFLICT (%s) DO UPDATE SET %s",
		r.tableName, strings.Join(insertFields, ", "), placeholders, conflictClause, updateClause)
}

// Upsert - UPSERT with ON CONFLICT
func (r *Repository[T]) Upsert(ctx context.Context, entity T, conflictCols []string, updateCols []string) error {
	query := r.BuildUpsertQuery(conflictCols, updateCols)

	_, err := r.db.NamedExecContext(ctx, query, entity)
	if err != nil {
		return fmt.Errorf("upsert failed: %w", err)
	}

	return nil
}
