package main

import "github.com/taskcluster/taskcluster-client-go/tcqueue"

func getTaskScopes(taskId string) ([]string, error) {
	// We do not need auth for this operation
	//
	// NOTE: currently, the only supported rootURL is https://taskcluster.net !!
	//
	// If this needs to change, remove this hardcoded reference, and make it a
	// (required) command line parameter to be passed to the proxy on start up.
	q := tcqueue.New(nil, "https://taskcluster.net")
	task, err := q.Task(taskId)
	if err != nil {
		return nil, err
	}

	return task.Scopes, nil
}
