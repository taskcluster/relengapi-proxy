package main

import (
	"github.com/taskcluster/taskcluster-client-go/queue"
)

func getTaskScopes(taskId string) ([]string, error) {
	// We do not need auth for this operation
	q := queue.NewNoAuth()

	task, err := q.Task(taskId)
	if err != nil {
		return nil, err
	}

	return task.Scopes, nil
}
