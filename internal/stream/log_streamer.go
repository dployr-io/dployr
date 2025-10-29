package stream

import (
	"bufio"
	"context"
	"dployr/pkg/core/utils"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type LogsStreamer struct {
	dir string
}

func Init() *LogsStreamer {
	homeDir, _ := os.UserHomeDir()
	return &LogsStreamer{
		dir: filepath.Join(homeDir, ".dployr", "logs"),
	}
}

type HandleLogStream interface {
	Stream(ctx context.Context, id string, logChan chan<- string) error
}

func (s *LogsStreamer) Stream(ctx context.Context, id string, logChan chan<- string) error {
	logPath := filepath.Join(s.dir, fmt.Sprintf("%s.log", utils.FormatName(id)))

	file, err := os.Open(logPath)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	
	// Send existing lines
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case logChan <- scanner.Text():
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading log file: %w", err)
	}

	// Follow the file for new lines
	stat, _ := file.Stat()
	lastSize := stat.Size()

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			stat, err := os.Stat(logPath)
			if err != nil {
				return fmt.Errorf("failed to stat log file: %w", err)
			}

			if stat.Size() > lastSize {
				file.Seek(lastSize, 0)
				scanner = bufio.NewScanner(file)
				
				for scanner.Scan() {
					select {
					case <-ctx.Done():
						return ctx.Err()
					case logChan <- scanner.Text():
					}
				}

				if err := scanner.Err(); err != nil {
					return fmt.Errorf("error reading new log lines: %w", err)
				}

				lastSize = stat.Size()
			}
		}
	}
}