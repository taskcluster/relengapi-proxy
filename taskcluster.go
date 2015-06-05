package main

import (
	queue "github.com/taskcluster/taskcluster-client-go/queue"
	"log"
)

func getTaskScopes(taskId string) []string {
	// We do not need auth for this operation
	q := queue.New("", "")
	q.Authenticate = false

	task, callSummary := q.Task(taskId)
	if callSummary.Error != nil {
		log.Fatalf("Could not fetch taskcluster task '%s' : %s",
			taskId, callSummary.Error)
	}

	return task.Scopes
}
