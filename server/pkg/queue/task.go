package queue

import (
	"context"
	"fmt"
)

type Task struct {
    Id         string        
    Payload    interface{}  
	Attempts   int           
    MaxRetries int   
    Errors     []error
    Messages   []string         
}

type Handler func(ctx context.Context, t *Task) error

func CreateHandler() Handler {
    return func(ctx context.Context, t *Task) error {
        fmt.Printf("handling %s (attempt %d)\n", t.Id, t.Attempts+1)
        // simulate transient failure twice
        if t.Attempts < 2 {
            return fmt.Errorf("transient error")
        }
        fmt.Printf("→ done %s\n", t.Id)
        return nil
    }
}