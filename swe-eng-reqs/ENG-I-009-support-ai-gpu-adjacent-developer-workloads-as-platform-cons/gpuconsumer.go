package main

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// Default chunk / artifact caps (Claude discretion per D-03 / plan).
// Per-chunk default 64KiB; total cap above MaxBytesReader 1<<20 for multi-chunk demos.
const (
	DefaultMaxChunkBytes = 64 * 1024       // 64KiB
	DefaultMaxTotalBytes = 8 * 1024 * 1024 // 8MiB
	DefaultJobTimeout    = 30 * time.Second
)

var (
	ErrJobNotFound     = errors.New("job not found")
	ErrJobTimedOut     = errors.New("job timed out")
	ErrChunkTooLarge   = errors.New("chunk exceeds max size")
	ErrTotalCapExceeded = errors.New("total artifact bytes exceeded")
)

// GPUJob is a long-running GPU-adjacent platform consumer job (ENG-I-009).
// Distinct from H-006 durable multi-step workflow Signal API.
type GPUJob struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	LongRunning    bool      `json:"long_running"`
	ChunkedUpload  bool      `json:"chunked_upload"`
	BytesReceived  int64     `json:"bytes_received"`
	Chunks         int       `json:"chunks"`
	Status         string    `json:"status"`
	TimedOut       bool      `json:"timed_out"`
	Deadline       time.Time `json:"deadline"`
	Timeout        string    `json:"timeout"`
	MaxChunkBytes  int       `json:"max_chunk_bytes"`
	MaxTotalBytes  int64     `json:"max_total_bytes"`
}

// GPUConsumer stores long-running jobs and accumulates chunked large-artifact uploads.
type GPUConsumer struct {
	mu            sync.Mutex
	jobs          map[string]*gpuJobInternal
	seq           int
	maxChunkBytes int
	maxTotalBytes int64
	now           func() time.Time
}

type gpuJobInternal struct {
	GPUJob
	buf []byte
}

// NewGPUConsumer creates a GPU-adjacent consumer store with default caps.
func NewGPUConsumer() *GPUConsumer {
	return &GPUConsumer{
		jobs:          make(map[string]*gpuJobInternal),
		maxChunkBytes: DefaultMaxChunkBytes,
		maxTotalBytes: DefaultMaxTotalBytes,
		now:           time.Now,
	}
}

// StartJob creates a long-running job with an absolute timeout deadline.
func (c *GPUConsumer) StartJob(name string, timeout time.Duration) (*GPUJob, error) {
	if name == "" {
		return nil, errors.New("name required")
	}
	if timeout <= 0 {
		timeout = DefaultJobTimeout
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.seq++
	id := fmt.Sprintf("job-%d", c.seq)
	deadline := c.now().Add(timeout)
	j := &gpuJobInternal{
		GPUJob: GPUJob{
			ID:            id,
			Name:          name,
			LongRunning:   true,
			ChunkedUpload: true,
			BytesReceived: 0,
			Chunks:        0,
			Status:        "running",
			TimedOut:      false,
			Deadline:      deadline,
			Timeout:       timeout.String(),
			MaxChunkBytes: c.maxChunkBytes,
			MaxTotalBytes: c.maxTotalBytes,
		},
	}
	c.jobs[id] = j
	return cloneJob(j), nil
}

// UploadChunk appends artifact bytes. Rejects oversize chunks and timed-out jobs (T-5-15).
func (c *GPUConsumer) UploadChunk(jobID string, data []byte, index int) (*GPUJob, error) {
	_ = index // ordering hint for clients; MVP appends in call order
	c.mu.Lock()
	defer c.mu.Unlock()
	j, ok := c.jobs[jobID]
	if !ok {
		return nil, ErrJobNotFound
	}
	if c.now().After(j.Deadline) {
		j.Status = "timeout"
		j.TimedOut = true
		return nil, ErrJobTimedOut
	}
	if len(data) > c.maxChunkBytes {
		return nil, ErrChunkTooLarge
	}
	next := j.BytesReceived + int64(len(data))
	if next > c.maxTotalBytes {
		return nil, ErrTotalCapExceeded
	}
	j.buf = append(j.buf, data...)
	j.BytesReceived = next
	j.Chunks++
	j.ChunkedUpload = true
	return cloneJob(j), nil
}

// GetJob returns a job snapshot, refreshing timeout status.
func (c *GPUConsumer) GetJob(jobID string) (*GPUJob, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	j, ok := c.jobs[jobID]
	if !ok {
		return nil, false
	}
	if !j.TimedOut && c.now().After(j.Deadline) {
		j.Status = "timeout"
		j.TimedOut = true
	}
	return cloneJob(j), true
}

func cloneJob(j *gpuJobInternal) *GPUJob {
	cp := j.GPUJob
	return &cp
}
