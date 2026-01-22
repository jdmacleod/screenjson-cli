package pipeline

import (
	"context"
	"time"

	"screenjson/cli/internal/model"
)

// JobStatus represents the status of a conversion job.
type JobStatus string

const (
	JobPending   JobStatus = "pending"
	JobRunning   JobStatus = "running"
	JobCompleted JobStatus = "completed"
	JobFailed    JobStatus = "failed"
)

// Job represents a conversion job.
type Job struct {
	ID           string
	Status       JobStatus
	InputFormat  string
	OutputFormat string
	InputPath    string
	OutputPath   string
	EncryptKey   string
	DecryptKey   string
	Created      time.Time
	Started      *time.Time
	Completed    *time.Time
	Error        error
	Document     *model.Document
	OutputData   []byte
}

// NewJob creates a new conversion job.
func NewJob(id string) *Job {
	return &Job{
		ID:      id,
		Status:  JobPending,
		Created: time.Now(),
	}
}

// Start marks the job as running.
func (j *Job) Start() {
	now := time.Now()
	j.Status = JobRunning
	j.Started = &now
}

// Complete marks the job as completed.
func (j *Job) Complete(doc *model.Document, output []byte) {
	now := time.Now()
	j.Status = JobCompleted
	j.Completed = &now
	j.Document = doc
	j.OutputData = output
}

// Fail marks the job as failed.
func (j *Job) Fail(err error) {
	now := time.Now()
	j.Status = JobFailed
	j.Completed = &now
	j.Error = err
}

// Duration returns the job duration.
func (j *Job) Duration() time.Duration {
	if j.Started == nil {
		return 0
	}
	end := time.Now()
	if j.Completed != nil {
		end = *j.Completed
	}
	return end.Sub(*j.Started)
}

// Decoder decodes a format into a ScreenJSON document.
type Decoder interface {
	Decode(ctx context.Context, data []byte) (*model.Document, error)
}

// Encoder encodes a ScreenJSON document into a format.
type Encoder interface {
	Encode(ctx context.Context, doc *model.Document) ([]byte, error)
}

// Codec combines Decoder and Encoder.
type Codec interface {
	Decoder
	Encoder
}
