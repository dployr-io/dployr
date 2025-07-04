package queue

import (
	"context"
	"log"

	"github.com/riverqueue/river"
)

type SetupUserJobArgs struct {
	BaseJobArgs
}

type SetupUserWorker struct {
	river.WorkerDefaults[SetupUserJobArgs]
}

func (args SetupUserJobArgs) Kind() string {
	return "setup_user"
}

func (args SetupUserJobArgs) GetBaseArgs() BaseJobArgs {
	return args.BaseJobArgs
}

type SetupUserData struct {
	UserID string `json:"user_id"`
}

func (w *SetupUserWorker) Work(ctx context.Context, job *river.Job[SetupUserJobArgs]) error {
	setupData, ok := job.Args.Data.(map[string]interface{})
	if !ok {
		return ErrInvalidJobData
	}

	userID, _ := setupData["user_id"].(string)

	log.Println("Setting up user", userID)

	return nil
}
