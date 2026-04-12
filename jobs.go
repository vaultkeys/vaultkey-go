package vaultkey

import "context"

// JobsService polls the status of async operations.
type JobsService struct {
	client *Client
}

// Get retrieves the current state of an async job.
//
// Most VaultKey operations that touch the chain are asynchronous and return
// a job ID. Poll this method until Status is JobStatusCompleted or JobStatusFailed.
//
//	result, apiErr, err := client.Jobs.Get(ctx, job.JobID)
//	if result.Status == vaultkey.JobStatusCompleted {
//	    fmt.Println("done:", result.Result)
//	}
func (j *JobsService) Get(ctx context.Context, jobID string) (Job, *ErrorResponse, error) {
	var resp Job
	apiErr, err := j.client.get(ctx, "/jobs/"+jobID, &resp)
	return resp, apiErr, err
}