package main

import "github.com/taskcluster/taskcluster-client-go/tcqueue"

func getTaskScopes(taskId string) ([]string, error) {
	// We do not need auth for this operation
	q := tcqueue.New(nil)
	task, err := q.Task(taskId)
	if err != nil {
		return nil, err
	}

	return task.Scopes, nil
}
