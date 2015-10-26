package main

import (
	queue "github.com/taskcluster/taskcluster-client-go/queue"
)

func getTaskScopes(taskId string) ([]string, error) {
	// We do not need auth for this operation
	q := queue.New("", "")
	q.Authenticate = false

	task, callSummary := q.Task(taskId)
	if callSummary.Error != nil {
		return nil, callSummary.Error
	}

	return task.Scopes, nil
}
