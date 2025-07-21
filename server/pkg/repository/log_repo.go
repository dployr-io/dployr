package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"dployr.io/pkg/models"
	"github.com/jmoiron/sqlx"
)


type LogRepo struct {
	*Repository[models.Event]
}

type LogFilters struct {
	ProjectId string     `json:"project_id"`
	Host      string     `json:"host"`
	Level     string     `json:"level"`
	Status    string     `json:"status"`
	Limit     int        `json:"limit"`
	Offset    int        `json:"offset"`
	StartDate *time.Time `json:"start_date"`
	EndDate   *time.Time `json:"end_date"`
}

type LogStream interface {
    Channel() <-chan *models.LogEntry // Returns a channel for log entries
    Close()                    // Closes the stream
}

type logStreamImpl struct {
    channel chan *models.LogEntry // Buffered channel for log entries
    done    chan bool            // Channel to signal closure
}

func (s *logStreamImpl) Channel() <-chan *models.LogEntry {
    return s.channel
}

func (s *logStreamImpl) Close() {
    close(s.done) // Signal closure
}

func NewLogRepo(db *sqlx.DB) *LogRepo {
	return &LogRepo{
		Repository: NewRepository[models.Event](db, "logs"),
	}
}

// GetWithFilters retrieves logs with applied filters and pagination
func (r *LogRepo) GetWithFilters(ctx context.Context, filters LogFilters) ([]models.LogEntry, int, error) {
	// Build the WHERE clause dynamically
	var conditions []string
	var args []interface{}
	argIndex := 1

	if filters.ProjectId != "" {
		conditions = append(conditions, fmt.Sprintf("project_id = ?%d", argIndex))
		args = append(args, filters.ProjectId)
		argIndex++
	}

	if filters.Host != "" {
		conditions = append(conditions, fmt.Sprintf("host = ?%d", argIndex))
		args = append(args, filters.Host)
		argIndex++
	}

	if filters.Level != "" {
		conditions = append(conditions, fmt.Sprintf("level = ?%d", argIndex))
		args = append(args, filters.Level)
		argIndex++
	}

	if filters.Status != "" {
		conditions = append(conditions, fmt.Sprintf("status = ?%d", argIndex))
		args = append(args, filters.Status)
		argIndex++
	}

	if filters.StartDate != nil {
		conditions = append(conditions, fmt.Sprintf("created_at >= ?%d", argIndex))
		args = append(args, filters.StartDate.Format(time.RFC3339))
		argIndex++
	}

	if filters.EndDate != nil {
		conditions = append(conditions, fmt.Sprintf("created_at <= ?%d", argIndex))
		args = append(args, filters.EndDate.Format(time.RFC3339))
		argIndex++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Get total count
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM logs %s", whereClause)
	var totalCount int
	if err := r.db.GetContext(ctx, &totalCount, countQuery, args...); err != nil {
		return nil, 0, fmt.Errorf("failed to get total count: %w", err)
	}

	// Get logs with pagination
	query := fmt.Sprintf(`
		SELECT id, project_id, host, message, created_at, level, status
		FROM logs 
		%s 
		ORDER BY created_at DESC 
		LIMIT ?%d OFFSET ?%d`,
		whereClause, argIndex, argIndex+1)

	args = append(args, filters.Limit, filters.Offset)

	var logs []models.LogEntry
	if err := r.db.SelectContext(ctx, &logs, query, args...); err != nil {
		return nil, 0, fmt.Errorf("failed to retrieve logs: %w", err)
	}

	return logs, totalCount, nil
}

// GetByID retrieves a specific log entry by ID
func (r *LogRepo) GetByID(ctx context.Context, id string) (models.LogEntry, error) {
	var log models.LogEntry
	query := `
		SELECT id, project_id, host, message, created_at, level, status
		FROM logs 
		WHERE id = ?`

	if err := r.db.GetContext(ctx, &log, query, id); err != nil {
		return log, fmt.Errorf("failed to retrieve log: %w", err)
	}

	return log, nil
}

// StreamLogs provides real-time log streaming (basic implementation)
func (r *LogRepo) StreamLogs(ctx context.Context, filters LogFilters) (LogStream, error) {
	stream := &logStreamImpl{
        channel: make(chan *models.LogEntry, 100), // Buffer for 100 log entries
        done:    make(chan bool),
    }

	// Start a goroutine to poll for new logs
	go func() {
		defer close(stream.channel)
		
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		
		var lastID string
		
		for {
			select {
			case <-ctx.Done():
				return
			case <-stream.done:
				return
			case <-ticker.C:
				// Get new logs since last check
				newLogs, err := r.getNewLogsSince(ctx, lastID, filters)
				if err != nil {
					continue
				}
				
				for _, log := range newLogs {
					select {
					case stream.channel <- &log:
						lastID = log.Id
					case <-ctx.Done():
						return
					case <-stream.done:
						return
					}
				}
			}
		}
	}()

	return stream, nil
}

// getNewLogsSince is a helper method for streaming
func (r *LogRepo) getNewLogsSince(ctx context.Context, lastID string, filters LogFilters) ([]models.LogEntry, error) {
	var conditions []string
	var args []interface{}
	argIndex := 1

	if lastID != "" {
		conditions = append(conditions, fmt.Sprintf("id > ?%d", argIndex))
		args = append(args, lastID)
		argIndex++
	}

	if filters.ProjectId != "" {
		conditions = append(conditions, fmt.Sprintf("project_id = ?%d", argIndex))
		args = append(args, filters.ProjectId)
		argIndex++
	}

	if filters.Host != "" {
		conditions = append(conditions, fmt.Sprintf("host = ?%d", argIndex))
		args = append(args, filters.Host)
		argIndex++
	}

	if filters.Level != "" {
		conditions = append(conditions, fmt.Sprintf("level = ?%d", argIndex))
		args = append(args, filters.Level)
		argIndex++
	}

	if filters.Status != "" {
		conditions = append(conditions, fmt.Sprintf("status = ?%d", argIndex))
		args = append(args, filters.Status)
		argIndex++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	query := fmt.Sprintf(`
		SELECT id, project_id, host, message, created_at, level, status
		FROM logs 
		%s 
		ORDER BY created_at ASC 
		LIMIT 50`, whereClause)

	var logs []models.LogEntry
	if err := r.db.SelectContext(ctx, &logs, query, args...); err != nil {
		return nil, fmt.Errorf("failed to retrieve new logs: %w", err)
	}

	return logs, nil
}
